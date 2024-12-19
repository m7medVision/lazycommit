import { $ } from "bun";
import { createOpenAI } from "@ai-sdk/openai";
import { createGoogleGenerativeAI } from "@ai-sdk/google";
import { generateText } from "ai";
import { readConfigFile } from "./config";
import simpleGit from "simple-git";

interface RunOptions {
  verbose?: boolean;
}

async function getStagedDiff(target_dir: string) {
  try {
    const git = simpleGit(target_dir);
    const diff = await git.diff(["--cached"]);
    return diff;
  } catch (error) {
    console.error("Error getting git diff:", error);
    throw error;
  }
}

export async function run(options: RunOptions, templateName?: string) {
  const config = await readConfigFile();
  if (options.verbose) {
    console.debug("Configuration loaded successfully.");
  }

  let templateFilePath: string;
  if (templateName) {
    if (!Object.prototype.hasOwnProperty.call(config.templates, templateName)) {
      console.error(
        `Error: Template '${templateName}' does not exist in the configuration.`
      );
      process.exit(1);
    }
    templateFilePath = config.templates[templateName];
    if (options.verbose) {
      console.debug(`Using template: ${templateName}`);
    }
  } else {
    templateFilePath = config.templates.default;
    if (options.verbose) {
      console.debug("Using default template.");
    }
  }

  const templateFile = Bun.file(templateFilePath);
  if (!(await templateFile.exists())) {
    console.error(
      `Error: The template file '${templateFilePath}' does not exist.`
    );
    process.exit(1);
  }
  if (options.verbose) {
    console.debug(`Template file found: ${templateFilePath}`);
  }

  const template = await templateFile.text();
  if (options.verbose) {
    console.debug("Template file read successfully.");
  }

  const target_dir = (await $`pwd`.text()).trim();
  if (options.verbose) {
    console.debug(`Target directory: ${target_dir}`);
  }

  if (!config.API_KEY) {
    console.error("API_KEY is not set");
    process.exit(1);
  }

  if (!config.model) {
    console.error("Model is not set");
    process.exit(1);
  }

  const diff = await getStagedDiff(target_dir);
  if (options.verbose) {
    console.debug("Git diff retrieved:\n", diff);
  }

  if (diff.trim().length === 0) {
    console.error(`No changes to commit in ${target_dir}`);
    process.exit(1);
  }

  const rendered_template = template.replace("{{diff}}", diff);
  if (options.verbose) {
    console.debug("Template rendered with git diff.");
  }

  const system_message =
    "You are a commit message generator. I will provide you with a git diff, and I would like you to generate an appropriate commit message using the conventional commit format. Do not write any explanations or other words, just reply with the commit message.";

  try {
    if (options.verbose) {
      console.debug("Sending request to AI service...");
    }

    let aiProvider;
    if (config.provider === "openai") {
      aiProvider = createOpenAI({
        compatibility: 'strict',
        apiKey: config.API_KEY,
      });
    } else if (config.provider === "google") {
      aiProvider = createGoogleGenerativeAI({
        apiKey: config.API_KEY,
      });
    } else {
      throw new Error("Invalid provider");
    }

    const { text } = await generateText({
      model: aiProvider(config.model),
      prompt: `${system_message}\n${rendered_template}`,
    });

    if (options.verbose) {
      console.debug("Response received from AI service.");
      console.debug(text);
    }

    console.log(text.trim());

    if (options.verbose) {
      console.debug("Commit message generated and outputted.");
    }

  } catch (error) {
    console.error(`Failed to fetch from AI service: ${error}`);
    process.exit(1);
  }
}