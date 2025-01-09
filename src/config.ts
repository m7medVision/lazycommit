import path from "path";
import os from "os";
import * as p from "@clack/prompts";
import OpenAI from "openai";
import { spawn } from "child_process";
import { template } from "./template";

async function editFile(filePath: string, onExit: () => void) {
	let editor =
		process.env.EDITOR ||
		(await p.select({
			message: "Select an editor",
			options: [
				{
					label: "neovim",
					value: "nvim",
				},
				{
					label: "vim",
					value: "vim",
				},
				{
					label: "nano",
					value: "nano",
				},
				{
					label: "code",
					value: "code",
				},
				{
					label: "cancel",
					value: "cancel",
				},
			],
		}));

	if (!editor || typeof editor !== "string" || editor === "cancel") {
		return;
	}

	let additionalArgs: string[] = [];
	if (/^(.[/\\])?code(.exe)?(\s+--.+)*/i.test(editor)) {
		editor = "code";
		additionalArgs = ["--wait"];
	}

	const child = spawn(editor, [filePath, ...additionalArgs], {
		stdio: "inherit",
	});

	await new Promise((resolve, reject) => {
		// biome-ignore lint/suspicious/noExplicitAny: unknown types to me
		child.on("exit", async (_e: any, _code: any) => {
			try {
				resolve(await onExit());
			} catch (error) {
				reject(error);
			}
		});
	});
}

function hasOwn<T extends object, K extends PropertyKey>(
	obj: T,
	key: K,
): obj is T & Record<K, unknown> {
	return key in obj && Object.prototype.hasOwnProperty.call(obj, key);
}

export const configPath = path.join(os.homedir(), ".lazycommit");

export interface Config {
	provider: "openai" | "google" | "custom";
	API_KEY: string;
	model: string;
	templates: Record<string, string>;
	customEndpoint?: string;
}

const DEFAULT_CONFIG: Config = {
	provider: "openai",
	API_KEY: "",
	model: "gpt-4",
	templates: {
		default: path.join(os.homedir(), ".lazycommit-template"),
	},
};

export async function readConfigFile(): Promise<Config> {
	const fileExists = await Bun.file(configPath).exists();
	if (!fileExists) {
		return DEFAULT_CONFIG;
	}

	const configString = await Bun.file(configPath).text();
	const config = JSON.parse(configString);

	return {
		...DEFAULT_CONFIG,
		...config,
	};
}

function validateKeys(keys: string[]): asserts keys is (keyof Config)[] {
	const configKeys = Object.keys(DEFAULT_CONFIG);

	for (const key of keys) {
		if (!configKeys.includes(key)) {
			throw new Error(`Invalid config property: ${key}`);
		}
	}
}

export async function cleanUpTemplates(config: Config): Promise<Config> {
	for (const templateName in config.templates) {
		const templatePath = config.templates[templateName];
		const fileExists = await Bun.file(templatePath).exists();
		if (!fileExists) {
			delete config.templates[templateName];
		}
	}
	return config;
}

export async function setConfigs(
	keyValues: [key: keyof Config, value: Config[keyof Config]][],
) {
	const config = await readConfigFile();

	validateKeys(keyValues.map(([key]) => key));

	for (const [key, value] of keyValues) {
		// @ts-ignore
		config[key] = value;
	}

	await Bun.write(configPath, JSON.stringify(config));
}

export async function showConfigUI() {
	try {
		const config = await cleanUpTemplates(await readConfigFile());

		const choice = (await p.select({
			message: "set config",
			options: [
				{
					label: "Provider",
					value: "provider",
					hint: config.provider,
				},
				{
					label: "API Key",
					value: "API_KEY",
					hint: hasOwn<Config, keyof Config>(config, "API_KEY")
						? `sk-...${config.API_KEY.slice(-3)}`
						: "not set",
				},
				{
					label: "Model",
					value: "model",
					hint: config.model,
				},
				{
					label: "Prompt Template",
					value: "template",
					hint: "edit the prompt template",
				},
				{
					label: "Custom Endpoint",
					value: "customEndpoint",
					hint: config.customEndpoint || "not set",
				},
				{
					label: "Done", // Changed from "Cancel" to "Done"
					value: "done",
					hint: "exit",
				},
			],
		})) as keyof Config | "template" | "done" | symbol;

		if (p.isCancel(choice) || choice === "done") {
			process.exit(0); // Ensure it exits when "Done" is selected
		}

		if (choice === "API_KEY") {
			const apiKey = await p.text({
				message: "API Key",
				initialValue: config.API_KEY,
			});

			await setConfigs([["API_KEY", apiKey as string]]);
		} else if (choice === "model") {
			const model = await p.select({
				message: "Model",
				options: (await getModels()).map((model) => ({
					label: model,
					value: model,
				})),
				initialValue: config.model,
			});

			await setConfigs([["model", model as string]]);
		} else if (choice === "template") {
			const templateChoice = (await p.select({
				message: "Choose a template to edit",
				options: [
					...Object.keys(config.templates).map((name) => ({
						label: name,
						value: name,
					})),
					{ label: "Add new template", value: "add_new" },
					{ label: "Cancel", value: "cancel" },
				],
			})) as string;

			if (templateChoice === "add_new") {
				const newTemplateName = (await p.text({
					message: "New template name",
				})) as string;

				const newTemplatePath = path.join(
					os.homedir(),
					`.lazycommit-template-${newTemplateName}`,
				);

				await Bun.write(newTemplatePath, template);
				config.templates[newTemplateName] = newTemplatePath;

				await editFile(newTemplatePath, async () => {
					console.log(`Prompt template '${newTemplateName}' updated`);
					await setConfigs([["templates", config.templates]]);
				});
			} else if (templateChoice !== "cancel") {
				const templatePath = config.templates[templateChoice];

				if (!(await Bun.file(templatePath).exists())) {
					await Bun.write(templatePath, template);
				}

				await editFile(templatePath, () => {
					console.log(`Prompt template '${templateChoice}' updated`);
				});
				process.exit(0);
			}
		} else if (choice === "provider") {
			const provider = await p.select({
				message: "Provider",
				options: [
					{ label: "OpenAI", value: "openai" },
					{ label: "Google", value: "google" },
					{ label: "Custom", value: "custom" },
				],
				initialValue: config.provider,
			});

			await setConfigs([["provider", provider as string]]);
		} else if (choice === "customEndpoint") {
			const customEndpoint = await p.text({
				message: "Custom Endpoint",
				initialValue: config.customEndpoint,
			});

			await setConfigs([["customEndpoint", customEndpoint as string]]);
		}

		if (p.isCancel(choice)) {
			process.exit(0);
		}

		showConfigUI();
		// biome-ignore lint/suspicious/noExplicitAny: unknown types to me
	} catch (error: any) {
		console.error(`\n${error.message}\n`);
	}
}

async function getModels() {
	const config = await readConfigFile();
	const provider = config.provider;
	const apiKey = config.API_KEY;
	if (!apiKey) {
		throw new Error("API_KEY is not set");
	}
	if (provider === "openai") {
		const oai = new OpenAI({
			apiKey,
			baseURL: config.customEndpoint,
		});
		const models = await oai.models.list();
		return models.data.map((model) => model.id);
	} else if (provider === "google") {
		return [
			"gemini-2.0-flash-exp",
			"gemini-1.5-pro-latest",
			"gemini-1.5-pro",
			"gemini-1.5-flash-latest",
			"gemini-1.5-flash",
		];
	}
	throw new Error("Invalid provider");
}
