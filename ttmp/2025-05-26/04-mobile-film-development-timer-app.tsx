import React, { useState, useEffect, useRef } from 'react';

const FilmDevelopmentTimer = () => {
  // Roll data from your routine
  const rollData = [
    { roll: 1, film: '120', pushPull: '+1 stop', devTime: '4:55', notes: 'Fresh chemistry—no fatigue compensation needed.' },
    { roll: 2, film: '120', pushPull: 'normal', devTime: '3:30', notes: 'First "tired" run; replenished.' },
    { roll: 3, film: '120', pushPull: 'normal', devTime: '3:36', notes: '' },
    { roll: 4, film: '120', pushPull: 'normal', devTime: '3:43', notes: '' },
    { roll: 5, film: '120', pushPull: 'normal', devTime: '3:49', notes: '' },
    { roll: 6, film: '120', pushPull: 'normal', devTime: '3:55', notes: '' },
    { roll: 7, film: '120', pushPull: 'normal', devTime: '4:01', notes: '' },
    { roll: 8, film: '35mm', pushPull: 'normal', devTime: '4:08', notes: '' },
    { roll: 9, film: '35mm', pushPull: 'normal', devTime: '4:14', notes: '' },
    { roll: 10, film: '35mm', pushPull: 'normal', devTime: '4:20', notes: '' },
    { roll: 11, film: '35mm', pushPull: 'normal', devTime: '4:27', notes: '' },
    { roll: 12, film: '35mm', pushPull: 'normal', devTime: '4:33', notes: '' },
    { roll: 13, film: '35mm', pushPull: 'normal', devTime: '4:39', notes: '' },
    { roll: 14, film: '35mm', pushPull: 'normal', devTime: '4:46', notes: '' },
    { roll: 15, film: '35mm', pushPull: 'normal', devTime: '4:52', notes: '' },
    { roll: 16, film: '35mm', pushPull: 'normal', devTime: '4:58', notes: '' },
    { roll: 17, film: '35mm', pushPull: 'normal', devTime: '5:04', notes: 'Final roll uses the last 25 mL from the reserve.' }
  ];

  const processSteps = [
    { name: 'Pre-soak', time: '1:00', description: 'Plain water to stabilize temperature' },
    { name: 'Developer', time: 'varies', description: 'Continuous gentle agitation first 10s of every 30s' },
    { name: 'Bleach-fix (Blix)', time: '6:30', description: 'No time change needed' },
    { name: 'Wash', time: '3:00', description: 'Running water 3-4 min' },
    { name: 'Stabilizer', time: '1:00', description: 'Then hang to dry' }
  ];

  const [currentRoll, setCurrentRoll] = useState(1);
  const [currentStep, setCurrentStep] = useState(0);
  const [timeLeft, setTimeLeft] = useState(0);
  const [isRunning, setIsRunning] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [completedSteps, setCompletedSteps] = useState([]);
  
  const intervalRef = useRef(null);
  const audioRef = useRef(null);

  // Convert time string to seconds
  const timeToSeconds = (timeStr) => {
    const [minutes, seconds] = timeStr.split(':').map(Number);
    return minutes * 60 + seconds;
  };

  // Convert seconds to time string (handles negative time)
  const secondsToTime = (seconds) => {
    const isNegative = seconds < 0;
    const absSeconds = Math.abs(seconds);
    const mins = Math.floor(absSeconds / 60);
    const secs = absSeconds % 60;
    const timeStr = `${mins}:${secs.toString().padStart(2, '0')}`;
    return isNegative ? `-${timeStr}` : timeStr;
  };

  // Get current step time
  const getCurrentStepTime = () => {
    if (currentStep === 1) { // Developer step
      return rollData[currentRoll - 1].devTime;
    }
    return processSteps[currentStep].time;
  };

  // Start timer
  const startTimer = () => {
    const stepTime = getCurrentStepTime();
    if (stepTime === 'varies') return;
    
    if (!isPaused) {
      setTimeLeft(timeToSeconds(stepTime));
    }
    setIsRunning(true);
    setIsPaused(false);
  };

  // Pause timer
  const pauseTimer = () => {
    setIsRunning(false);
    setIsPaused(true);
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
    }
  };

  // Stop timer
  const stopTimer = () => {
    setIsRunning(false);
    setIsPaused(false);
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
    }
    setTimeLeft(0);
  };

  // Select step by clicking
  const selectStep = (stepIndex) => {
    stopTimer();
    setCurrentStep(stepIndex);
    setTimeLeft(0);
  };

  // Complete current step
  const completeStep = () => {
    stopTimer();
    setCompletedSteps([...completedSteps, currentStep]);
    if (currentStep < processSteps.length - 1) {
      setCurrentStep(currentStep + 1);
    }
    setTimeLeft(0);
  };

  // Reset all steps for current roll
  const resetSteps = () => {
    stopTimer();
    setCurrentStep(0);
    setCompletedSteps([]);
    setTimeLeft(0);
    setIsPaused(false);
  };

  // Next roll
  const nextRoll = () => {
    if (currentRoll < 17) {
      setCurrentRoll(currentRoll + 1);
      resetSteps();
    }
  };

  // Previous roll
  const prevRoll = () => {
    if (currentRoll > 1) {
      setCurrentRoll(currentRoll - 1);
      resetSteps();
    }
  };

  // Timer effect
  useEffect(() => {
    if (isRunning) {
      intervalRef.current = setInterval(() => {
        setTimeLeft((prev) => {
          const newTime = prev - 1;
          // Play notification sound only when crossing from positive to negative
          if (prev === 1 && audioRef.current) {
            audioRef.current.play().catch(() => {});
          }
          return newTime;
        });
      }, 1000);
    } else {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [isRunning]);

  const currentRollData = rollData[currentRoll - 1];

  return (
    <div className="min-h-screen bg-gray-900 text-white p-4">
      <audio ref={audioRef} preload="auto">
        <source src="data:audio/wav;base64,UklGRnoGAABXQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YQoGAACBhYqFbF1fdJivrJBhNjVgodDbq2EcBj+a2/LDciUFLIHO8tiJNwgZaLvt559NEAxQp+PwtmMcBjiR1/LMeSwFJHfH8N2QQAoUXrTp66hVFApGn+DyvGUeCSuBzvLZizoIGGq+7eGdTgwPVKzn77BdGwU+ltryxnkpBSl+0fPaizsIGGq+7eCdUAwSU6ng8bllHgU5j9n0wnMnBSZ+0fPaizwIF2m98OKcTQwPU6vj8bJfHAU8l9j1yHUpBSZ/0/Pai8AII2y68d+STAwNSqXe9bdkHAU+mdkAAA" />
      </audio>

      <div className="max-w-md mx-auto">
        {/* Header */}
        <div className="text-center mb-6">
          <h1 className="text-2xl font-bold mb-2">🎞️ Film Development Timer</h1>
          <p className="text-gray-400">Cinestill Cs41 • 102°F/39°C</p>
        </div>

        {/* Roll Selector */}
        <div className="bg-gray-800 rounded-lg p-4 mb-6">
          <div className="flex items-center justify-between mb-4">
            <button 
              onClick={prevRoll}
              disabled={currentRoll === 1}
              className="bg-blue-600 disabled:bg-gray-600 px-4 py-2 rounded-lg font-semibold"
            >
              ← Prev
            </button>
            <div className="text-center">
              <div className="text-lg font-bold">Roll {currentRoll}/17</div>
              <div className="text-sm text-gray-400">{currentRollData.film} • {currentRollData.pushPull}</div>
            </div>
            <button 
              onClick={nextRoll}
              disabled={currentRoll === 17}
              className="bg-blue-600 disabled:bg-gray-600 px-4 py-2 rounded-lg font-semibold"
            >
              Next →
            </button>
          </div>
          
          {currentRollData.notes && (
            <div className="text-sm text-yellow-400 bg-yellow-900/20 p-2 rounded">
              💡 {currentRollData.notes}
            </div>
          )}
        </div>

        {/* Process Steps */}
        <div className="space-y-3 mb-6">
          {processSteps.map((step, index) => (
            <div 
              key={index}
              onClick={() => selectStep(index)}
              className={`p-4 rounded-lg border-2 cursor-pointer transition-all ${
                index === currentStep 
                  ? 'border-blue-500 bg-blue-900/20' 
                  : completedSteps.includes(index)
                  ? 'border-green-500 bg-green-900/20'
                  : 'border-gray-600 bg-gray-800 hover:border-gray-500'
              }`}
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className="text-lg">
                    {completedSteps.includes(index) ? '✅' : index === currentStep ? '⏱️' : '⏸️'}
                  </span>
                  <span className="font-semibold">{step.name}</span>
                </div>
                <span className="font-mono text-lg">
                  {index === 1 ? currentRollData.devTime : step.time}
                </span>
              </div>
              <p className="text-sm text-gray-400">{step.description}</p>
            </div>
          ))}
        </div>

        {/* Timer Display */}
        <div className="bg-gray-800 rounded-lg p-6 mb-6 text-center">
          <div className={`text-4xl font-mono font-bold mb-4 ${timeLeft < 0 ? 'text-red-400' : 'text-white'}`}>
            {secondsToTime(timeLeft)}
          </div>
          {timeLeft < 0 && (
            <div className="text-red-400 text-sm mb-4">⚠️ OVERTIME</div>
          )}
          
          <div className="flex gap-2 justify-center flex-wrap">
            {!isRunning && !isPaused && (
              <button 
                onClick={startTimer}
                disabled={getCurrentStepTime() === 'varies'}
                className="bg-green-600 disabled:bg-gray-600 px-4 py-2 rounded-lg font-semibold flex items-center gap-2"
              >
                ▶️ Start
              </button>
            )}
            
            {isPaused && (
              <button 
                onClick={startTimer}
                className="bg-green-600 px-4 py-2 rounded-lg font-semibold flex items-center gap-2"
              >
                ▶️ Resume
              </button>
            )}
            
            {isRunning && (
              <button 
                onClick={pauseTimer}
                className="bg-yellow-600 px-4 py-2 rounded-lg font-semibold flex items-center gap-2"
              >
                ⏸️ Pause
              </button>
            )}
            
            <button 
              onClick={stopTimer}
              disabled={!isRunning && !isPaused}
              className="bg-red-600 disabled:bg-gray-600 px-4 py-2 rounded-lg font-semibold flex items-center gap-2"
            >
              ⏹️ Stop
            </button>
            
            <button 
              onClick={completeStep}
              className="bg-blue-600 px-4 py-2 rounded-lg font-semibold flex items-center gap-2"
            >
              ✅ Complete
            </button>
          </div>
        </div>

        {/* Quick Actions */}
        <div className="flex gap-3 mb-6">
          <button 
            onClick={resetSteps}
            className="flex-1 bg-yellow-600 px-4 py-3 rounded-lg font-semibold"
          >
            🔄 Reset Roll
          </button>
        </div>

        {/* Quick Reference */}
        <div className="bg-gray-800 rounded-lg p-4">
          <h3 className="font-bold mb-2">📋 Quick Checklist</h3>
          <div className="text-sm space-y-1 text-gray-300">
            <div>1. Pre-soak (1 min) - stabilize temperature</div>
            <div>2. Developer - gentle agitation first 10s of every 30s</div>
            <div>3. Bleach-fix (6:30) - no time change needed</div>
            <div>4. Wash (3-4 min) - running water</div>
            <div>5. Stabilizer (1 min) - then hang to dry</div>
          </div>
        </div>

        {/* Replenishment Reminder */}
        {currentStep === processSteps.length - 1 && (
          <div className="mt-4 bg-orange-900/20 border border-orange-500 rounded-lg p-4">
            <h3 className="font-bold text-orange-400 mb-2">🧪 After This Roll:</h3>
            <div className="text-sm text-orange-300">
              <div>• Pour used 600mL back to WORKING bottle</div>
              <div>• Bleed off 25mL (≈4%) to waste</div>
              <div>• Replace with 25mL fresh from RESERVE</div>
              <div>• Cap both bottles, invert once to mix</div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default FilmDevelopmentTimer;