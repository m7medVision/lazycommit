#!/usr/bin/env bun
import { cli, command } from "cleye";
import { version } from "../package.json";
import { setConfigs, showConfigUI, type Config } from "./config";
import { run } from "./run";
import { initialize } from "./utils";

const config = command(
	{
		name: "config",
		help: {
			description: "Configure lazycommit",
		},
	},
	(argv) => {
		(async () => {
			const [mode, ...keyValues] = argv._;
			await initialize();

			if (!mode) {
				await showConfigUI();
				return;
			}

			if (!keyValues.length) {
				console.error(`Error: Missing required parameter "key=value"\n`);
				argv.showHelp();
				return process.exit(1);
			}

			if (mode === "set") {
				if (keyValues.includes("templates")) {
					console.error("Error: Templates not settable");
					return process.exit(1);
				}

				await setConfigs(
					keyValues.map((kv: string) => {
						const [key, value] = kv.split("=");
						return [key, value] as [keyof Config, string];
					}),
				);

				return;
			}

			console.error(`Error: Invalid mode: ${mode}`);
			process.exit(1);
		})();
	},
);

export const CLI = cli(
	{
		name: "lazycommit",
		version,
		commands: [config],
		flags: {
			template: String,
			verbose: Boolean,
		},
	},
	(argv) => {
		(async () => {
			process.on("unhandledRejection", (reason, promise) => {
				console.error(
					"Unhandled Rejection at:",
					promise,
					"reason:",
					reason,
					"\nPlease report this! https://github.com/m7medvision/lazycommit/issues",
				);
				process.exit(1);
			});

			process.on("uncaughtException", (err) => {
				console.error(
					"Unhandled exception. \nPlease report this! https://github.com/m7medvision/lazycommit/issues",
					err,
				);
				process.exit(1);
			});

			const { template, verbose } = argv.flags;

			if (verbose) {
				console.debug("Verbose mode enabled");
				console.debug("Flags:", argv.flags);
			}

			await initialize();

			await run({ verbose }, template);
		})();
	},
);
