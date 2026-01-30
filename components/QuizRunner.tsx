
import React, { useState, useEffect, useRef } from 'react';
import { QuizQuestion, QuizResult, QuizMode, QuestionPerformance } from '../types';

interface Props {
  questions: QuizQuestion[];
  mode: QuizMode;
  onComplete: (result: QuizResult) => void;
}

const QuizRunner: React.FC<Props> = ({ questions, mode, onComplete }) => {
  const [currentIndex, setCurrentIndex] = useState(0);
  const [selectedAnswers, setSelectedAnswers] = useState<(number | null)[]>(new Array(questions.length).fill(null));
  const [timeLeft, setTimeLeft] = useState<number>(0);
  const [performances, setPerformances] = useState<QuestionPerformance[]>(
    questions.map(q => ({
      questionId: q.id,
      timeSpentSeconds: 0,
      attempts: 0,
      isCorrect: false,
      revisits: 0
    }))
  );

  const startTimeRef = useRef<number>(Date.now());
  const questionStartTimeRef = useRef<number>(Date.now());
  const visitCounts = useRef<number[]>(new Array(questions.length).fill(0));

  useEffect(() => {
    // Initialize first question
    visitCounts.current[0] = 1;
    if (mode === QuizMode.TIMED_QUESTION) {
      setTimeLeft(questions[0].timeLimitSeconds);
    } else {
      const totalTime = questions.reduce((acc, q) => acc + q.timeLimitSeconds, 0);
      setTimeLeft(totalTime);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const timer = setInterval(() => {
      setTimeLeft(prev => {
        if (prev <= 1) {
          if (mode === QuizMode.TIMED_QUESTION) {
            handleNext();
            return 0;
          }
          handleFinish();
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [currentIndex, mode]);

  const recordCurrentPerformance = () => {
    const timeSpent = (Date.now() - questionStartTimeRef.current) / 1000;
    setPerformances(prev => {
      const updated = [...prev];
      updated[currentIndex] = {
        ...updated[currentIndex],
        timeSpentSeconds: updated[currentIndex].timeSpentSeconds + timeSpent,
        revisits: visitCounts.current[currentIndex] - 1
      };
      return updated;
    });
  };

  const handleNext = () => {
    recordCurrentPerformance();
    if (currentIndex < questions.length - 1) {
      const nextIdx = currentIndex + 1;
      setCurrentIndex(nextIdx);
      visitCounts.current[nextIdx] += 1;
      questionStartTimeRef.current = Date.now();
      if (mode === QuizMode.TIMED_QUESTION) {
        setTimeLeft(questions[nextIdx].timeLimitSeconds);
      }
    } else {
      handleFinish();
    }
  };

  const handlePrev = () => {
    recordCurrentPerformance();
    if (currentIndex > 0) {
      const nextIdx = currentIndex - 1;
      setCurrentIndex(nextIdx);
      visitCounts.current[nextIdx] += 1;
      questionStartTimeRef.current = Date.now();
      // Per-question timer resets if we go back? 
      // Pedagogically better to keep it ticking or reset? 
      // User requirements say "individual time limit", let's reset it for the revisit.
      if (mode === QuizMode.TIMED_QUESTION) {
        setTimeLeft(questions[nextIdx].timeLimitSeconds);
      }
    }
  };

  const handleSelect = (idx: number) => {
    const newAnswers = [...selectedAnswers];
    newAnswers[currentIndex] = idx;
    setSelectedAnswers(newAnswers);
    
    setPerformances(prev => {
      const updated = [...prev];
      updated[currentIndex] = {
        ...updated[currentIndex],
        attempts: updated[currentIndex].attempts + 1,
        isCorrect: idx === questions[currentIndex].correctIndex
      };
      return updated;
    });
  };

  const handleFinish = () => {
    recordCurrentPerformance();
    const totalTimeSpent = (Date.now() - startTimeRef.current) / 1000;
    const correctCount = selectedAnswers.reduce((acc, ans, idx) => {
      return acc + (ans === questions[idx].correctIndex ? 1 : 0);
    }, 0);

    onComplete({
      totalQuestions: questions.length,
      correctAnswers: correctCount,
      totalTimeSpent,
      performances,
      mode
    });
  };

  const progress = ((currentIndex + 1) / questions.length) * 100;
  const currentQuestion = questions[currentIndex];

  return (
    <div className="space-y-6">
      {/* Progress & Header */}
      <div className="bg-white p-4 rounded-2xl shadow-sm border border-slate-100 flex items-center justify-between">
        <div className="flex-1 mr-8">
          <div className="flex justify-between text-xs font-bold text-slate-400 mb-2 uppercase tracking-widest">
            <span>Question {currentIndex + 1} of {questions.length}</span>
            <span>{Math.round(progress)}%</span>
          </div>
          <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
            <div 
              className="h-full bg-indigo-600 transition-all duration-500 ease-out" 
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
        <div className={`flex items-center gap-2 px-4 py-2 rounded-xl border ${timeLeft < 10 ? 'bg-red-50 border-red-200 text-red-600' : 'bg-slate-50 border-slate-200 text-slate-700'}`}>
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span className="font-mono font-bold text-lg leading-none">{timeLeft}s</span>
        </div>
      </div>

      {/* Question Card */}
      <div className="bg-white p-6 sm:p-10 rounded-2xl shadow-xl shadow-slate-200/50 border border-slate-100 animate-in fade-in slide-in-from-bottom-4 duration-500">
        <h3 className="text-xl sm:text-2xl font-semibold text-slate-800 mb-8 leading-relaxed">
          {currentQuestion.text}
        </h3>

        <div className="grid grid-cols-1 gap-4">
          {currentQuestion.options.map((option, idx) => {
            const isSelected = selectedAnswers[currentIndex] === idx;
            return (
              <button
                key={idx}
                onClick={() => handleSelect(idx)}
                className={`group flex items-center text-left p-5 rounded-xl border-2 transition-all ${
                  isSelected 
                    ? 'border-indigo-600 bg-indigo-50 shadow-md ring-2 ring-indigo-200' 
                    : 'border-slate-100 hover:border-indigo-200 hover:bg-slate-50'
                }`}
              >
                <div className={`w-8 h-8 rounded-full flex items-center justify-center mr-4 transition-colors ${
                  isSelected ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-500 group-hover:bg-indigo-100 group-hover:text-indigo-600'
                }`}>
                  {String.fromCharCode(65 + idx)}
                </div>
                <span className={`flex-1 font-medium ${isSelected ? 'text-indigo-900' : 'text-slate-600'}`}>
                  {option}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {/* Navigation */}
      <div className="flex items-center justify-between">
        <button
          onClick={handlePrev}
          disabled={currentIndex === 0}
          className="flex items-center gap-2 px-6 py-3 rounded-xl border-2 border-slate-200 text-slate-600 font-bold hover:bg-slate-50 transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 19l-7-7 7-7" /></svg>
          Back
        </button>
        
        {currentIndex === questions.length - 1 ? (
          <button
            onClick={handleFinish}
            disabled={selectedAnswers[currentIndex] === null}
            className="flex items-center gap-2 px-8 py-3 rounded-xl bg-indigo-600 text-white font-bold hover:bg-indigo-700 shadow-lg shadow-indigo-200 transition-all disabled:opacity-50"
          >
            Submit Quiz
          </button>
        ) : (
          <button
            onClick={handleNext}
            disabled={selectedAnswers[currentIndex] === null}
            className="flex items-center gap-2 px-8 py-3 rounded-xl bg-indigo-600 text-white font-bold hover:bg-indigo-700 shadow-lg shadow-indigo-200 transition-all disabled:opacity-50"
          >
            Next
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" /></svg>
          </button>
        )}
      </div>
    </div>
  );
};

export default QuizRunner;
