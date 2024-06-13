import { $ } from "bun";
import OpenAI from "openai";
import { readConfigFile } from "./config";
import simpleGit from "simple-git";
import { GoogleGenerativeAI } from "@google/generative-ai";

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
    throw error; // Re-throw the error after logging it
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
        `Error: Template '${templateName}' does not exist in the configuration.`,
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
      `Error: The template file '${templateFilePath}' does not exist.`,
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
  if (config.provider === "openai") {
    const oai = new OpenAI({
      apiKey: config.API_KEY,
    });

    try {
      if (options.verbose) {
        console.debug("Sending request to OpenAI...");
      }
      const response = await oai.chat.completions.create({
        messages: [
          {
            role: "system",
            content: system_message,
          },
          {
            role: "user",
            content: rendered_template,
          },
        ],
        model: config.model,
        response_format: { type: "json_object" },
      });

      if (options.verbose) {
        console.debug("Response received from OpenAI.");
        console.debug(JSON.stringify(response, null, 2));
      }

      const content = response.choices[0].message.content;
      if (!content) {
        console.error("Failed to generate commit message");
        process.exit(1);
      }
      try {
        const content_json = JSON.parse(content);
        for (const message of content_json.commitMessages) {
          console.log(message);
        }
      }
      catch (error) {
        console.error("Error parsing JSON response:", error);
        process.exit(1);
      }
      if (options.verbose) {
        console.debug("Commit message generated and outputted.");
      }
    } catch (error) {
      console.error(`Failed to fetch from openai: ${error}`);
      process.exit(1);
    }
  } else if (config.provider === "google") {
    const genAI = new GoogleGenerativeAI(config.API_KEY);
    const model = genAI.getGenerativeModel({
      model: config.model,
      systemInstruction: system_message,
      generationConfig: { responseMimeType: "application/json" },
    });
    const session = model.startChat({
      history: [],
    });
    const response = await session.sendMessage(rendered_template);
    try {
      const content_json = JSON.parse(response.response.text());
      for (const message of content_json.commitMessages) {
        console.log(message);
      }
    } catch (error) {
      console.error("Error parsing JSON response:", error);
      process.exit(1);
    }
  } else {
    console.error("Provider not supported");
    process.exit(1);
  }
}
