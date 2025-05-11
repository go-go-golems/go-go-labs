import React, { useState, useEffect } from 'react';
import { Clock, Code, MessagesSquare, Users, Info, Github, Check, List, ArrowRight } from 'lucide-react';

const StreamInfoDisplay = () => {
  // State to store all the stream information
  const [streamInfo, setStreamInfo] = useState({
    title: "Building a React Component Library",
    description: "Creating reusable UI components with TailwindCSS",
    startTime: new Date().toISOString(),
    language: "JavaScript/React",
    githubRepo: "https://github.com/yourusername/component-library",
    currentTask: "",
    viewerCount: 42,
  });

  // State for steps (completed, current, upcoming)
  const [completedSteps, setCompletedSteps] = useState([
    "Project setup and initialization",
    "Design system planning"
  ]);
  
  const [activeStep, setActiveStep] = useState("Setting up component architecture");
  
  const [upcomingSteps, setUpcomingSteps] = useState([
    "Implement Button component",
    "Create Card component",
    "Build Form elements",
    "Add dark mode toggle"
  ]);

  // State for new step input
  const [newStep, setNewStep] = useState("");
  const [newTopic, setNewTopic] = useState("");

  // State for editing mode
  const [isEditing, setIsEditing] = useState(false);
  const [editableInfo, setEditableInfo] = useState({...streamInfo});
  
  // Calculate stream duration
  const [duration, setDuration] = useState("00:00:00");
  
  useEffect(() => {
    const updateDuration = () => {
      const start = new Date(streamInfo.startTime);
      const now = new Date();
      const diff = Math.floor((now - start) / 1000);
      
      const hours = Math.floor(diff / 3600).toString().padStart(2, '0');
      const minutes = Math.floor((diff % 3600) / 60).toString().padStart(2, '0');
      const seconds = (diff % 60).toString().padStart(2, '0');
      
      setDuration(`${hours}:${minutes}:${seconds}`);
    };
    
    updateDuration();
    const interval = setInterval(updateDuration, 1000);
    
    return () => clearInterval(interval);
  }, [streamInfo.startTime]);
  
  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setEditableInfo(prev => ({...prev, [name]: value}));
  };
  
  const saveChanges = () => {
    setStreamInfo({...editableInfo});
    setIsEditing(false);
  };
  
  const cancelChanges = () => {
    setEditableInfo({...streamInfo});
    setIsEditing(false);
  };
  
  const resetTimer = () => {
    const newInfo = {...streamInfo, startTime: new Date().toISOString()};
    setStreamInfo(newInfo);
    setEditableInfo(newInfo);
  };

  const addNewStep = () => {
    if (newStep.trim()) {
      setUpcomingSteps([...upcomingSteps, newStep.trim()]);
      setNewStep("");
    }
  };

  const setNewActiveTopic = () => {
    if (newTopic.trim()) {
      if (activeStep) {
        setCompletedSteps([...completedSteps, activeStep]);
      }
      setActiveStep(newTopic.trim());
      setNewTopic("");
    }
  };

  const completeCurrentStep = () => {
    if (activeStep) {
      setCompletedSteps([...completedSteps, activeStep]);
      if (upcomingSteps.length > 0) {
        setActiveStep(upcomingSteps[0]);
        setUpcomingSteps(upcomingSteps.slice(1));
      } else {
        setActiveStep("");
      }
    }
  };

  const makeStepActive = (step, source) => {
    if (activeStep) {
      setCompletedSteps([...completedSteps, activeStep]);
    }
    
    setActiveStep(step);
    
    if (source === 'upcoming') {
      setUpcomingSteps(upcomingSteps.filter(s => s !== step));
    } else if (source === 'completed') {
      setCompletedSteps(completedSteps.filter(s => s !== step));
    }
  };

  return (
    <div className="w-full max-w-4xl mx-auto p-6 bg-white text-black rounded-none shadow-lg font-mono" style={{fontFamily: 'monospace'}}>
      <div className="border-b-2 border-black pb-4 mb-6">
        <div className="flex justify-between items-center">
          <h1 className="text-2xl font-bold uppercase tracking-widest">LUMON INDUSTRIES</h1>
          <div className="flex items-center">
            <div className="flex flex-col items-end mr-6">
              <div className="text-xs uppercase">MACRODATA STREAM</div>
              <div className="text-xl font-bold">{duration}</div>
            </div>
            {!isEditing ? (
              <button 
                onClick={() => setIsEditing(true)}
                className="px-4 py-2 bg-black text-white rounded-none hover:bg-gray-800 transition-colors uppercase text-xs tracking-wider"
              >
                Edit Parameters
              </button>
            ) : (
              <>
                <button 
                  onClick={saveChanges}
                  className="px-4 py-2 bg-green-900 text-white rounded-none hover:bg-green-800 transition-colors uppercase text-xs tracking-wider mr-2"
                >
                  Save
                </button>
                <button 
                  onClick={cancelChanges}
                  className="px-4 py-2 bg-red-900 text-white rounded-none hover:bg-red-800 transition-colors uppercase text-xs tracking-wider"
                >
                  Cancel
                </button>
              </>
            )}
          </div>
        </div>
      </div>
      
      {isEditing ? (
        <div className="grid grid-cols-1 gap-4 p-4 border-2 border-black">
          <div>
            <label className="block text-sm font-medium mb-1 uppercase tracking-wider">Stream Title</label>
            <input
              name="title"
              value={editableInfo.title}
              onChange={handleInputChange}
              className="w-full p-2 border-2 border-black rounded-none bg-white text-black"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium mb-1 uppercase tracking-wider">Description</label>
            <textarea
              name="description"
              value={editableInfo.description}
              onChange={handleInputChange}
              rows="2"
              className="w-full p-2 border-2 border-black rounded-none bg-white text-black"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium mb-1 uppercase tracking-wider">Programming Language/Framework</label>
            <input
              name="language"
              value={editableInfo.language}
              onChange={handleInputChange}
              className="w-full p-2 border-2 border-black rounded-none bg-white text-black"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium mb-1 uppercase tracking-wider">GitHub Repository</label>
            <input
              name="githubRepo"
              value={editableInfo.githubRepo}
              onChange={handleInputChange}
              className="w-full p-2 border-2 border-black rounded-none bg-white text-black"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium mb-1 uppercase tracking-wider">Viewer Count</label>
            <input
              type="number"
              name="viewerCount"
              value={editableInfo.viewerCount}
              onChange={handleInputChange}
              className="w-full p-2 border-2 border-black rounded-none bg-white text-black"
            />
          </div>
          
          <div className="flex items-center">
            <button
              onClick={resetTimer}
              className="px-4 py-2 bg-black text-white rounded-none hover:bg-gray-800 transition-colors uppercase text-xs tracking-wider"
            >
              Reset Timer
            </button>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          <div className="lg:col-span-5 border-2 border-black p-4">
            <div className="mb-6">
              <div className="text-xs uppercase tracking-wider mb-1">PROJECT DESIGNATION</div>
              <h2 className="font-bold text-lg">{streamInfo.title}</h2>
              <p className="text-sm mt-1">{streamInfo.description}</p>
            </div>
            
            <div className="flex items-center mb-3">
              <div className="w-6 h-6 mr-2 flex items-center justify-center bg-black text-white">
                <Code size={16} />
              </div>
              <div>
                <div className="text-xs uppercase tracking-wider">LANGUAGE</div>
                <span>{streamInfo.language}</span>
              </div>
            </div>
            
            <div className="flex items-center mb-3">
              <div className="w-6 h-6 mr-2 flex items-center justify-center bg-black text-white">
                <Github size={16} />
              </div>
              <div>
                <div className="text-xs uppercase tracking-wider">REPOSITORY</div>
                <a 
                  href={streamInfo.githubRepo} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="text-blue-900 hover:underline break-all"
                >
                  {streamInfo.githubRepo}
                </a>
              </div>
            </div>
            
            <div className="flex items-center">
              <div className="w-6 h-6 mr-2 flex items-center justify-center bg-black text-white">
                <Users size={16} />
              </div>
              <div>
                <div className="text-xs uppercase tracking-wider">VIEWERS</div>
                <span>{streamInfo.viewerCount}</span>
              </div>
            </div>
            
            <div className="mt-6">
              <div className="flex">
                <input
                  type="text"
                  value={newTopic}
                  onChange={(e) => setNewTopic(e.target.value)}
                  placeholder="Enter new topic..."
                  className="flex-grow p-2 border-2 border-black rounded-none bg-white text-black"
                />
                <button
                  onClick={setNewActiveTopic}
                  className="px-4 py-2 bg-black text-white rounded-none hover:bg-gray-800 transition-colors uppercase text-xs tracking-wider"
                >
                  Set Topic
                </button>
              </div>
            </div>
          </div>
          
          <div className="lg:col-span-7 border-2 border-black">
            <div className="border-b-2 border-black p-3 bg-black text-white">
              <h3 className="uppercase tracking-wider font-bold">CURRENT PROGRESS</h3>
            </div>
            
            {/* Active step */}
            <div className="p-4 border-b-2 border-black bg-green-50">
              <div className="flex justify-between items-center">
                <div>
                  <div className="text-xs uppercase tracking-wider mb-1 text-green-800">ACTIVE TASK</div>
                  <p className="font-bold">{activeStep || "No active task"}</p>
                </div>
                {activeStep && (
                  <button
                    onClick={completeCurrentStep}
                    className="px-3 py-1 bg-green-900 text-white rounded-none hover:bg-green-800 transition-colors uppercase text-xs tracking-wider flex items-center"
                  >
                    <Check size={14} className="mr-1" /> Complete
                  </button>
                )}
              </div>
            </div>
            
            {/* Completed steps */}
            <div className="p-4 border-b-2 border-black">
              <div className="text-xs uppercase tracking-wider mb-2">COMPLETED TASKS ({completedSteps.length})</div>
              {completedSteps.length > 0 ? (
                <ul className="divide-y divide-gray-200">
                  {completedSteps.map((step, index) => (
                    <li key={index} className="py-2 flex justify-between items-center">
                      <div className="flex items-center">
                        <div className="w-4 h-4 mr-2 bg-black text-white flex items-center justify-center">
                          <Check size={12} />
                        </div>
                        <span className="line-through">{step}</span>
                      </div>
                      <button
                        onClick={() => makeStepActive(step, 'completed')}
                        className="text-xs text-blue-900 hover:underline"
                      >
                        Reactivate
                      </button>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="text-sm text-gray-500 italic">No completed tasks yet</p>
              )}
            </div>
            
            {/* Upcoming steps */}
            <div className="p-4">
              <div className="text-xs uppercase tracking-wider mb-2">UPCOMING TASKS ({upcomingSteps.length})</div>
              {upcomingSteps.length > 0 ? (
                <ul className="divide-y divide-gray-200">
                  {upcomingSteps.map((step, index) => (
                    <li key={index} className="py-2 flex justify-between items-center">
                      <div className="flex items-center">
                        <div className="w-4 h-4 mr-2 border border-black flex items-center justify-center">
                          {index === 0 ? <ArrowRight size={12} /> : (index + 1)}
                        </div>
                        <span>{step}</span>
                      </div>
                      <button
                        onClick={() => makeStepActive(step, 'upcoming')}
                        className="text-xs text-blue-900 hover:underline"
                      >
                        Make Active
                      </button>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="text-sm text-gray-500 italic">No upcoming tasks</p>
              )}
              
              <div className="mt-4 flex">
                <input
                  type="text"
                  value={newStep}
                  onChange={(e) => setNewStep(e.target.value)}
                  placeholder="Add new step..."
                  className="flex-grow p-2 border-2 border-black rounded-none bg-white text-black"
                />
                <button
                  onClick={addNewStep}
                  className="px-4 py-2 bg-black text-white rounded-none hover:bg-gray-800 transition-colors uppercase text-xs tracking-wider"
                >
                  Add
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default StreamInfoDisplay;