
import React, { useState } from 'react';

interface Props {
  onLogin: (email: string) => void;
}

const AuthForm: React.FC<Props> = ({ onLogin }) => {
  const [isLogin, setIsLogin] = useState(true);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    // Simulate API call
    setTimeout(() => {
      onLogin(email);
      setIsLoading(false);
    }, 1000);
  };

  return (
    <div className="bg-white/70 backdrop-blur-xl rounded-3xl shadow-2xl shadow-indigo-100/50 p-8 sm:p-12 max-w-md mx-auto border border-white">
      <div className="text-center mb-10">
        <div className="w-16 h-16 bg-indigo-600 rounded-2xl flex items-center justify-center mx-auto mb-6 shadow-lg shadow-indigo-200">
          <span className="text-white font-bold text-3xl">EP</span>
        </div>
        <h2 className="text-3xl font-black text-slate-900 mb-2">
          {isLogin ? 'Welcome Back' : 'Create Account'}
        </h2>
        <p className="text-slate-500 font-medium">
          {isLogin ? 'Log in to continue your journey' : 'Join the future of personalized learning'}
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-5">
        <div className="space-y-2">
          <label className="text-sm font-bold text-slate-700 ml-1">Email Address</label>
          <input
            required
            type="email"
            className="w-full px-5 py-4 rounded-2xl border border-slate-200 focus:ring-4 focus:ring-indigo-100 focus:border-indigo-500 transition-all outline-none bg-white/50"
            placeholder="student@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-bold text-slate-700 ml-1">Password</label>
          <input
            required
            type="password"
            className="w-full px-5 py-4 rounded-2xl border border-slate-200 focus:ring-4 focus:ring-indigo-100 focus:border-indigo-500 transition-all outline-none bg-white/50"
            placeholder="••••••••"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>

        <button
          disabled={isLoading}
          type="submit"
          className="w-full py-5 bg-indigo-600 hover:bg-indigo-700 text-white font-bold rounded-2xl shadow-xl shadow-indigo-200 transition-all transform active:scale-95 disabled:opacity-70 flex items-center justify-center gap-3"
        >
          {isLoading ? (
            <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
          ) : (
            isLogin ? 'Sign In' : 'Sign Up'
          )}
        </button>
      </form>

      <div className="mt-8 text-center">
        <button
          onClick={() => setIsLogin(!isLogin)}
          className="text-sm font-bold text-indigo-600 hover:text-indigo-800 transition-colors"
        >
          {isLogin ? "Don't have an account? Sign Up" : "Already have an account? Sign In"}
        </button>
      </div>
    </div>
  );
};

export default AuthForm;
