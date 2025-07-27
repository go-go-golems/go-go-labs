// Initialize global counter (safe for re-execution)
if (!globalState.counter) {
    globalState.counter = 0;
}

// Basic routes using Express.js API
app.get("/", (req, res) => {
    res.send("JS playground online with Geppetto APIs");
});

app.get("/health", (req, res) => {
    res.json({ok: true, counter: globalState.counter});
});

app.post("/counter", (req, res) => {
    res.json({count: ++globalState.counter});
});

// Example Geppetto API usage route
app.get("/geppetto-demo", (req, res) => {
    try {
        // Create a new conversation
        const conv = new Conversation();
        
        // Add a simple message (note: using lowercase method names due to field name mapper)
        const msgId = conv.addMessage("user", "Hello, Geppetto!");
        console.log("Added message with ID:", msgId);
        
        // Get conversation as single prompt
        const prompt = conv.getSinglePrompt();
        
        res.json({
            success: true,
            messageId: msgId,
            prompt: prompt,
            conversationAPI: "Available",
            chatFactory: typeof ChatStepFactory !== 'undefined' ? "Available" : "Not Available"
        });
    } catch (error) {
        console.error("Geppetto demo error:", error);
        res.status(500).json({
            success: false,
            error: error.message
        });
    }
});

console.log("Bootstrap complete - server ready with Geppetto APIs");