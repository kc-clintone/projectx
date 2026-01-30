
import React, { useState, useEffect } from 'react';
import { UserProfile, QuizMode, QuizQuestion, QuizResult, AIAnalysis } from './types';
import ProfileForm from './components/ProfileForm';
import QuizConfig from './components/QuizConfig';
import QuizRunner from './components/QuizRunner';
import AnalysisDashboard from './components/AnalysisDashboard';
import AuthForm from './components/AuthForm';
import { generateQuiz, analyzePerformance } from './services/geminiService';
import { db } from './services/db';

const App: React.FC = () => {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [quizQuestions, setQuizQuestions] = useState<QuizQuestion[] | null>(null);
  const [currentResult, setCurrentResult] = useState<QuizResult | null>(null);
  const [analysis, setAnalysis] = useState<AIAnalysis | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [activeStep, setActiveStep] = useState<'profile' | 'config' | 'quiz' | 'analysis'>('profile');

  useEffect(() => {
    const user = db.getUser();
    if (user) {
      setIsLoggedIn(true);
      const savedProfile = db.getProfile();
      if (savedProfile) {
        setProfile(savedProfile);
        setActiveStep('config');
      } else {
        setActiveStep('profile');
      }
    }
  }, []);

  const handleLogin = (email: string) => {
    db.saveUser(email);
    setIsLoggedIn(true);
    // Check if profile exists immediately after login
    const savedProfile = db.getProfile();
    if (savedProfile) {
      setProfile(savedProfile);
      setActiveStep('config');
    } else {
      setActiveStep('profile');
    }
  };

  const handleLogout = () => {
    db.logout();
    setIsLoggedIn(false);
    setProfile(null);
    setActiveStep('profile');
  };

  const handleProfileSubmit = (data: UserProfile) => {
    db.saveProfile(data);
    setProfile(data);
    setActiveStep('config');
  };

  const handleStartQuiz = async (topic: string, mode: QuizMode, image?: string) => {
    if (!profile) return;
    setIsLoading(true);
    try {
      const response = await generateQuiz(profile, topic, mode, image);
      
      if (!response.isValid) {
        alert(`Content Policy Alert: ${response.reason || "This topic does not meet our educational guidelines. Please provide an academic or learning-focused topic."}`);
        return;
      }

      setQuizQuestions(response.questions);
      setActiveStep('quiz');
    } catch (error) {
      console.error("Failed to generate quiz", error);
      alert("Error generating quiz. Please check your connection and try again.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleQuizComplete = async (result: QuizResult) => {
    if (!profile || !quizQuestions) return;
    db.saveQuizResult(result);
    setCurrentResult(result);
    setIsLoading(true);
    setActiveStep('analysis');
    try {
      const analysisData = await analyzePerformance(profile, result, quizQuestions);
      setAnalysis(analysisData);
    } catch (error) {
      console.error("Failed to analyze performance", error);
    } finally {
      setIsLoading(false);
    }
  };

  const resetQuiz = () => {
    setQuizQuestions(null);
    setCurrentResult(null);
    setAnalysis(null);
    setActiveStep('config');
  };

  if (!isLoggedIn) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center p-4 bg-[radial-gradient(circle_at_top_right,_var(--tw-gradient-stops))] from-indigo-50 via-slate-50 to-violet-50">
        <AuthForm onLogin={handleLogin} />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-50 flex flex-col">
      <header className="bg-white/80 backdrop-blur-md border-b border-slate-200 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-20 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-indigo-600 rounded-2xl flex items-center justify-center shadow-lg shadow-indigo-100">
              <span className="text-white font-bold text-2xl">EP</span>
            </div>
            <div>
              <h1 className="text-xl font-black bg-gradient-to-r from-indigo-600 to-violet-600 bg-clip-text text-transparent leading-none">
                EduPulse AI
              </h1>
              <span className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Personalized Learning</span>
            </div>
          </div>
          <div className="flex items-center gap-6">
            {profile && activeStep !== 'profile' && (
              <button 
                onClick={() => setActiveStep('profile')}
                className="text-sm font-bold text-slate-600 hover:text-indigo-600 transition-colors hidden sm:block"
              >
                Hi, {profile.name}
              </button>
            )}
            <button 
              onClick={handleLogout}
              className="px-5 py-2.5 rounded-xl border border-slate-200 text-sm font-bold text-slate-600 hover:bg-slate-50 hover:text-red-600 transition-all"
            >
              Sign Out
            </button>
          </div>
        </div>
      </header>

      <main className="flex-1 max-w-5xl mx-auto w-full p-4 sm:p-8">
        {isLoading && (
          <div className="fixed inset-0 bg-white/60 backdrop-blur-sm z-[60] flex flex-col items-center justify-center gap-6 text-center px-4">
            <div className="relative">
               <div className="w-20 h-20 border-4 border-indigo-100 rounded-full"></div>
               <div className="w-20 h-20 border-4 border-indigo-600 border-t-transparent rounded-full animate-spin absolute top-0"></div>
            </div>
            <div className="space-y-2">
              <p className="text-indigo-900 font-bold animate-pulse text-xl">
                {activeStep === 'quiz' ? 'AI is crafting your questions...' : 'Synthesizing your performance data...'}
              </p>
              <p className="text-slate-500 text-sm max-w-xs mx-auto">Ensuring content aligns with educational guidelines.</p>
            </div>
          </div>
        )}

        <div className="animate-in fade-in slide-in-from-bottom-8 duration-700">
          {activeStep === 'profile' && <ProfileForm onSubmit={handleProfileSubmit} initialData={profile} />}
          {activeStep === 'config' && <QuizConfig onStart={handleStartQuiz} />}
          {activeStep === 'quiz' && quizQuestions && (
            <QuizRunner 
              questions={quizQuestions} 
              mode={QuizMode.TIMED_QUESTION}
              onComplete={handleQuizComplete} 
            />
          )}
          {activeStep === 'analysis' && currentResult && (
            <AnalysisDashboard 
              result={currentResult} 
              analysis={analysis} 
              questions={quizQuestions || []}
              onReset={resetQuiz}
            />
          )}
        </div>
      </main>

      <footer className="bg-white border-t border-slate-100 py-10">
        <div className="max-w-7xl mx-auto px-4 text-center">
          <p className="text-slate-400 text-sm font-medium">
            &copy; {new Date().getFullYear()} EduPulse AI. An intelligent agent platform.
          </p>
        </div>
      </footer>
    </div>
  );
};

export default App;
