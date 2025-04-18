import express, { Request, Response } from "express";
import OpenAI from "openai";
import dotenv from "dotenv";
import cors from "cors"; // Import cors

dotenv.config();

const app = express();
const port = process.env.PORT || 3001;

// Use cors middleware
app.use(cors()); // Allow all origins for development

app.use(express.json());

// Initialize OpenAI client (ensure OPENAI_API_KEY is in your .env)
const openai = new OpenAI({
  apiKey: process.env.OPENAI_API_KEY,
});

// Define the chat endpoint
app.post("/api/chat", async (req: Request, res: Response) => {
  const { prompt, currentText } = req.body;

  if (!prompt || typeof currentText !== "string") {
    return res.status(400).json({ error: "Missing prompt or currentText" });
  }

  try {
    const completion = await openai.chat.completions.create({
      model: "gpt-4o-mini", // Or your preferred model
      messages: [
        {
          role: "system",
          content:
            "You are a helpful writing assistant. You will be given the current text in a textarea and an edit request. Your goal is to apply the edit request to the current text and return ONLY the full, revised text, without any extra explanations, commentary, or conversational filler. If the request is ambiguous or cannot be applied, ask for clarification.",
        },
        {
          role: "user",
          content: `Here is the current text:\n\n${currentText}`,
        },
        {
          role: "user",
          content: `Please apply this edit request: ${prompt}`,
        },
      ],
      temperature: 0.7, // Adjust as needed
    });

    const revisedText = completion.choices[0]?.message?.content?.trim();

    if (revisedText === undefined) {
      console.error("OpenAI response was empty or malformed:", completion);
      return res.status(500).json({ error: "Failed to get response from AI" });
    }

    // Return both the revised text and the same text as 'assistant' response for the chat
    res.json({ revisedText, assistant: revisedText });
  } catch (error) {
    console.error("Error calling OpenAI:", error);
    res.status(500).json({ error: "Failed to process chat request" });
  }
});

app.listen(port, () => {
  console.log(`Server listening on port ${port}`);
});
