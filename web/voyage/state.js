class State {
    constructor() {
        this.loadFromLocalStorage();
    }

    loadFromLocalStorage() {
        const savedState = localStorage.getItem('midjourneyPromptState');
        if (savedState) {
            this.data = JSON.parse(savedState);

            // make sure all the values are present, initialize with default if not
            if (!this.data.prompt_fragments) {
                this.data.prompt_fragments = [];
            }
            if (!this.data.images) {
                this.data.images = [];
            }
            if (!this.data.options) {
                this.data.options = {
                    aspect_ratio: "16:9",
                    model_version: "v5"
                };
            }  
            if (!this.data.current_prompt) {
                this.data.current_prompt = "";
            }
            if (!this.data.prompt_history) {
                this.data.prompt_history = [ ];
            }
            if (!this.data.search_query) {
                this.data.search_query = "";
            }
            if (!this.data.checked_fragments) {
                this.save();
            }
            if (!this.data.saved_selections) {
                this.data.saved_selections = [];    
            }
                this.save();
        } else {
            this.data = {
                prompt_fragments: [
                    "a majestic lion",
                    "in a lush jungle",
                    "with vibrant colors",
                    "photorealistic style"
                ],
                images: [
                    { url: "https://example.com/lion.jpg", thumbnail: "", alt: "Lion" },
                    { url: "https://example.com/jungle.jpg", thumbnail: "", alt: "Jungle" }
                ],
                options: {
                    aspect_ratio: "16:9",
                    model_version: "v5"
                },
                current_prompt: "",
                prompt_history: [
                    "a serene lake at sunset",
                    "cyberpunk cityscape with neon lights",
                    "abstract geometric patterns in pastel colors"
                ],
                search_query: "",
                checked_fragments: [],
                saved_selections: []
            };
            this.save();
        }
    }

    get(key) {
        return this.data[key];
    }

    set(key, value) {
        this.data[key] = value;
        this.save();
    }

    save() {
        localStorage.setItem('midjourneyPromptState', JSON.stringify(this.data));
    }

    addToHistory(prompt) {
        if (prompt && prompt !== this.data.prompt_history[0]) {
            this.data.prompt_history.unshift(prompt);
            this.save();
        }
    }
}

export default State;