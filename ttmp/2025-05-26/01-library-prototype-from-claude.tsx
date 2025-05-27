import React, { useState, useMemo } from 'react';

const ChatHistoryBrowser = () => {
  // Mock data
  const [chats, setChats] = useState([
    {
      id: 1,
      title: "JavaScript Array Methods",
      date: "2024-05-25",
      model: "Claude Sonnet 4",
      folderId: 1,
      tags: ["javascript", "arrays", "programming"],
      notes: "Great explanation of map, filter, and reduce",
      favorite: true,
      preview: "Can you explain the difference between map and forEach in JavaScript?"
    },
    {
      id: 2,
      title: "React Best Practices",
      date: "2024-05-24",
      model: "Claude Sonnet 4",
      folderId: 1,
      tags: ["react", "best-practices", "frontend"],
      notes: "Useful patterns for component design",
      favorite: false,
      preview: "What are some React best practices for large applications?"
    },
    {
      id: 3,
      title: "Machine Learning Basics",
      date: "2024-05-23",
      model: "Claude Opus 4",
      folderId: 2,
      tags: ["ml", "basics", "education"],
      notes: "Good introduction to ML concepts",
      favorite: true,
      preview: "Explain machine learning in simple terms"
    },
    {
      id: 4,
      title: "Python Data Analysis",
      date: "2024-05-22",
      model: "Claude Sonnet 4",
      folderId: 2,
      tags: ["python", "data-analysis", "pandas"],
      notes: "Pandas tutorial with examples",
      favorite: false,
      preview: "How do I analyze CSV data using pandas?"
    },
    {
      id: 5,
      title: "Creative Writing Tips",
      date: "2024-05-21",
      model: "Claude Opus 4",
      folderId: 3,
      tags: ["writing", "creative", "tips"],
      notes: "",
      favorite: false,
      preview: "Give me some creative writing prompts for short stories"
    }
  ]);

  const [folders, setFolders] = useState([
    { id: 1, name: "Programming", color: "bg-blue-500", count: 2 },
    { id: 2, name: "Learning", color: "bg-green-500", count: 2 },
    { id: 3, name: "Creative", color: "bg-purple-500", count: 1 }
  ]);

  const [allTags] = useState([
    "javascript", "react", "python", "ml", "arrays", "programming", 
    "best-practices", "frontend", "basics", "education", "data-analysis", 
    "pandas", "writing", "creative", "tips"
  ]);

  // State
  const [selectedChat, setSelectedChat] = useState(null);
  const [selectedFolder, setSelectedFolder] = useState(null);
  const [selectedTags, setSelectedTags] = useState([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [showFavorites, setShowFavorites] = useState(false);
  const [sortBy, setSortBy] = useState("date");
  const [editingNotes, setEditingNotes] = useState(null);
  const [newNote, setNewNote] = useState("");

  // Filtered chats
  const filteredChats = useMemo(() => {
    let filtered = chats;

    // Filter by folder
    if (selectedFolder) {
      filtered = filtered.filter(chat => chat.folderId === selectedFolder);
    }

    // Filter by tags
    if (selectedTags.length > 0) {
      filtered = filtered.filter(chat => 
        selectedTags.every(tag => chat.tags.includes(tag))
      );
    }

    // Filter by search query
    if (searchQuery) {
      filtered = filtered.filter(chat =>
        chat.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
        chat.preview.toLowerCase().includes(searchQuery.toLowerCase()) ||
        chat.notes.toLowerCase().includes(searchQuery.toLowerCase()) ||
        chat.tags.some(tag => tag.toLowerCase().includes(searchQuery.toLowerCase()))
      );
    }

    // Filter by favorites
    if (showFavorites) {
      filtered = filtered.filter(chat => chat.favorite);
    }

    // Sort
    filtered.sort((a, b) => {
      switch (sortBy) {
        case "date":
          return new Date(b.date) - new Date(a.date);
        case "title":
          return a.title.localeCompare(b.title);
        case "model":
          return a.model.localeCompare(b.model);
        default:
          return 0;
      }
    });

    return filtered;
  }, [chats, selectedFolder, selectedTags, searchQuery, showFavorites, sortBy]);

  const toggleFavorite = (chatId) => {
    setChats(chats.map(chat => 
      chat.id === chatId ? { ...chat, favorite: !chat.favorite } : chat
    ));
  };

  const updateNotes = (chatId, notes) => {
    setChats(chats.map(chat => 
      chat.id === chatId ? { ...chat, notes } : chat
    ));
    setEditingNotes(null);
    setNewNote("");
  };

  const toggleTag = (tag) => {
    setSelectedTags(prev => 
      prev.includes(tag) 
        ? prev.filter(t => t !== tag)
        : [...prev, tag]
    );
  };

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Sidebar */}
      <div className="w-80 bg-white border-r border-gray-200 flex flex-col">
        {/* Header */}
        <div className="p-4 border-b border-gray-200">
          <h1 className="text-xl font-bold text-gray-900 mb-4">📚 Chat Library</h1>
          
          {/* Search */}
          <div className="relative mb-4">
            <input
              type="text"
              placeholder="Search chats..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <span className="absolute left-3 top-2.5 text-gray-400">🔍</span>
          </div>

          {/* Controls */}
          <div className="flex gap-2 mb-4">
            <button
              onClick={() => setShowFavorites(!showFavorites)}
              className={`px-3 py-1 rounded-full text-sm ${
                showFavorites 
                  ? 'bg-yellow-100 text-yellow-800' 
                  : 'bg-gray-100 text-gray-600'
              }`}
            >
              ⭐ Favorites
            </button>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value)}
              className="px-3 py-1 rounded-lg text-sm border border-gray-300"
            >
              <option value="date">📅 Date</option>
              <option value="title">📝 Title</option>
              <option value="model">🤖 Model</option>
            </select>
          </div>
        </div>

        {/* Folders */}
        <div className="p-4 border-b border-gray-200">
          <h3 className="font-semibold text-gray-700 mb-3">📁 Folders</h3>
          <div className="space-y-2">
            <button
              onClick={() => setSelectedFolder(null)}
              className={`w-full text-left px-3 py-2 rounded-lg transition-colors ${
                selectedFolder === null 
                  ? 'bg-blue-100 text-blue-800' 
                  : 'hover:bg-gray-100'
              }`}
            >
              📂 All Chats ({chats.length})
            </button>
            {folders.map(folder => (
              <button
                key={folder.id}
                onClick={() => setSelectedFolder(folder.id)}
                className={`w-full text-left px-3 py-2 rounded-lg transition-colors ${
                  selectedFolder === folder.id 
                    ? 'bg-blue-100 text-blue-800' 
                    : 'hover:bg-gray-100'
                }`}
              >
                <span className={`inline-block w-3 h-3 rounded-full ${folder.color} mr-2`}></span>
                {folder.name} ({folder.count})
              </button>
            ))}
          </div>
        </div>

        {/* Tags */}
        <div className="p-4 flex-1 overflow-y-auto">
          <h3 className="font-semibold text-gray-700 mb-3">🏷️ Tags</h3>
          <div className="flex flex-wrap gap-2">
            {allTags.map(tag => (
              <button
                key={tag}
                onClick={() => toggleTag(tag)}
                className={`px-2 py-1 rounded-full text-xs transition-colors ${
                  selectedTags.includes(tag)
                    ? 'bg-blue-500 text-white'
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                }`}
              >
                #{tag}
              </button>
            ))}
          </div>
          
          {selectedTags.length > 0 && (
            <div className="mt-4">
              <button
                onClick={() => setSelectedTags([])}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                🗑️ Clear filters
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex">
        {/* Chat List */}
        <div className="w-96 bg-white border-r border-gray-200">
          <div className="p-4 border-b border-gray-200">
            <h2 className="font-semibold text-gray-900">
              Chats ({filteredChats.length})
            </h2>
          </div>
          
          <div className="overflow-y-auto h-full">
            {filteredChats.map(chat => (
              <div
                key={chat.id}
                onClick={() => setSelectedChat(chat)}
                className={`p-4 border-b border-gray-100 cursor-pointer transition-colors ${
                  selectedChat?.id === chat.id 
                    ? 'bg-blue-50 border-blue-200' 
                    : 'hover:bg-gray-50'
                }`}
              >
                <div className="flex items-start justify-between mb-2">
                  <h3 className="font-medium text-gray-900 text-sm">{chat.title}</h3>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      toggleFavorite(chat.id);
                    }}
                    className="text-lg"
                  >
                    {chat.favorite ? '⭐' : '☆'}
                  </button>
                </div>
                
                <p className="text-xs text-gray-600 mb-2">{chat.preview}</p>
                
                <div className="flex items-center justify-between text-xs text-gray-500">
                  <span>{chat.date}</span>
                  <span className="px-2 py-1 bg-gray-100 rounded">{chat.model}</span>
                </div>
                
                <div className="flex flex-wrap gap-1 mt-2">
                  {chat.tags.slice(0, 3).map(tag => (
                    <span
                      key={tag}
                      className="px-2 py-0.5 bg-blue-100 text-blue-700 rounded-full text-xs"
                    >
                      #{tag}
                    </span>
                  ))}
                  {chat.tags.length > 3 && (
                    <span className="px-2 py-0.5 bg-gray-100 text-gray-500 rounded-full text-xs">
                      +{chat.tags.length - 3}
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Chat Detail */}
        <div className="flex-1 bg-white">
          {selectedChat ? (
            <div className="h-full flex flex-col">
              {/* Chat Header */}
              <div className="p-6 border-b border-gray-200">
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h1 className="text-2xl font-bold text-gray-900 mb-2">{selectedChat.title}</h1>
                    <div className="flex items-center gap-4 text-sm text-gray-600">
                      <span>📅 {selectedChat.date}</span>
                      <span>🤖 {selectedChat.model}</span>
                      <span>📁 {folders.find(f => f.id === selectedChat.folderId)?.name}</span>
                    </div>
                  </div>
                  <button
                    onClick={() => toggleFavorite(selectedChat.id)}
                    className="text-2xl hover:scale-110 transition-transform"
                  >
                    {selectedChat.favorite ? '⭐' : '☆'}
                  </button>
                </div>

                {/* Tags */}
                <div className="flex flex-wrap gap-2 mb-4">
                  {selectedChat.tags.map(tag => (
                    <span
                      key={tag}
                      className="px-3 py-1 bg-blue-100 text-blue-700 rounded-full text-sm"
                    >
                      #{tag}
                    </span>
                  ))}
                </div>

                {/* Notes Section */}
                <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium text-gray-900">📝 Notes</h3>
                    {editingNotes !== selectedChat.id && (
                      <button
                        onClick={() => {
                          setEditingNotes(selectedChat.id);
                          setNewNote(selectedChat.notes);
                        }}
                        className="text-sm text-blue-600 hover:text-blue-800"
                      >
                        ✏️ Edit
                      </button>
                    )}
                  </div>
                  
                  {editingNotes === selectedChat.id ? (
                    <div>
                      <textarea
                        value={newNote}
                        onChange={(e) => setNewNote(e.target.value)}
                        placeholder="Add your notes about this chat..."
                        className="w-full p-3 border border-gray-300 rounded-lg resize-none focus:ring-2 focus:ring-blue-500"
                        rows={3}
                      />
                      <div className="flex gap-2 mt-2">
                        <button
                          onClick={() => updateNotes(selectedChat.id, newNote)}
                          className="px-3 py-1 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700"
                        >
                          💾 Save
                        </button>
                        <button
                          onClick={() => {
                            setEditingNotes(null);
                            setNewNote("");
                          }}
                          className="px-3 py-1 bg-gray-300 text-gray-700 rounded-lg text-sm hover:bg-gray-400"
                        >
                          ❌ Cancel
                        </button>
                      </div>
                    </div>
                  ) : (
                    <p className="text-gray-700">
                      {selectedChat.notes || "No notes yet. Click edit to add some!"}
                    </p>
                  )}
                </div>
              </div>

              {/* Chat Content */}
              <div className="flex-1 p-6 bg-gray-50">
                <div className="bg-white rounded-lg p-6 shadow-sm">
                  <h3 className="font-medium text-gray-900 mb-4">💬 Chat Preview</h3>
                  <div className="space-y-4">
                    <div className="bg-blue-50 border-l-4 border-blue-500 p-4 rounded">
                      <p className="text-sm text-gray-600 mb-1">👤 You</p>
                      <p className="text-gray-900">{selectedChat.preview}</p>
                    </div>
                    <div className="bg-green-50 border-l-4 border-green-500 p-4 rounded">
                      <p className="text-sm text-gray-600 mb-1">🤖 {selectedChat.model}</p>
                      <p className="text-gray-900">This is where the AI response would be displayed. In a real implementation, you would load the full chat history and display the conversation here with proper formatting and syntax highlighting.</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          ) : (
            <div className="h-full flex items-center justify-center text-gray-500">
              <div className="text-center">
                <div className="text-6xl mb-4">💭</div>
                <h2 className="text-xl font-medium mb-2">Select a Chat</h2>
                <p>Choose a conversation from the list to view its details</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ChatHistoryBrowser;