# 🎭 Emoji Agent Example

This example demonstrates how to customize AISH's AI prompt to create a specialized agent that responds only with emojis.

## 🚀 Quick Start

Run AISH in this directory:

```bash
aish
```

## ⚙️ How It Works

The `.aishrc` file in this directory is automatically read and sourced by AISH when you start the shell. In this configuration, we use the `aiprompt` built-in command to instruct the LLM model to respond using **emojis only**.

### Configuration Details

See the [`.aishrc`](.aishrc) file for the complete configuration:

```bash
# Custom system prompt to make AI respond with emojis only
aiprompt "You are an emoji-only agent. Always respond with emojis and never use text."
```

## 🎯 Example Output

When you interact with AISH in this directory, the AI will respond using emojis:

![Emoji Agent Demo](./emoji.png)

## 🔧 Try It Out

Once you're in AISH with this configuration loaded, try asking questions like:

- "How are you feeling today?"
- "What's the weather like?"
- "Tell me a joke"

The AI will respond with expressive emojis instead of text!

## 📁 Files

- **`.aishrc`** - Configuration file that sets up the emoji-only prompt
- **`emoji.png`** - Screenshot showing the emoji agent in action
- **`README.md`** - This documentation file

---

<div align="center">
  <p><em>Create your own specialized AI agents with custom prompts!</em></p>
</div>
