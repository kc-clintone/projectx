
import React, { useState } from 'react';
import { QuizMode } from '../types';

interface Props {
  onStart: (topic: string, mode: QuizMode, image?: string) => void;
}

const QuizConfig: React.FC<Props> = ({ onStart }) => {
  const [topic, setTopic] = useState('');
  const [mode, setMode] = useState<QuizMode>(QuizMode.TIMED_QUESTION);
  const [image, setImage] = useState<string | undefined>();
  const [preview, setPreview] = useState<string | null>(null);

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onloadend = () => {
        const result = reader.result as string;
        setImage(result);
        setPreview(result);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!topic.trim()) return;
    onStart(topic, mode, image);
  };

  return (
    <div className="bg-white rounded-2xl shadow-xl shadow-slate-200/50 p-6 sm:p-10 max-w-2xl mx-auto border border-slate-100">
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-slate-900 mb-2">Create New Quiz</h2>
        <p className="text-slate-500 text-sm">Input a topic or upload your study material to begin.</p>
        <div className="mt-4 p-3 bg-amber-50 border border-amber-100 rounded-xl flex items-start gap-3">
          <svg className="w-5 h-5 text-amber-500 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
          </svg>
          <p className="text-[11px] text-amber-800 font-medium leading-relaxed">
            <strong>Educational Policy:</strong> EduPulse AI is strictly for academic and learning purposes. Topics involving violence, illegal acts, or inappropriate content will be automatically rejected.
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="space-y-2">
          <label className="text-sm font-semibold text-slate-700">Study Topic</label>
          <input
            required
            className="w-full px-4 py-3 rounded-xl border border-slate-200 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all outline-none"
            placeholder="e.g. Molecular Biology, Calculus, Modern History"
            value={topic}
            onChange={e => setTopic(e.target.value)}
          />
        </div>

        <div className="space-y-2">
          <label className="text-sm font-semibold text-slate-700">Optional: Study Material</label>
          <div className="flex items-center justify-center w-full">
            <label className="flex flex-col items-center justify-center w-full h-32 border-2 border-slate-200 border-dashed rounded-xl cursor-pointer bg-slate-50 hover:bg-slate-100 transition-colors">
              <div className="flex flex-col items-center justify-center pt-5 pb-6">
                <svg className="w-8 h-8 mb-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                </svg>
                <p className="mb-2 text-sm text-slate-500">
                  <span className="font-semibold">Click to upload</span> or drag and drop
                </p>
                <p className="text-xs text-slate-400">Notes, Diagrams, Textbooks (MAX. 5MB)</p>
              </div>
              <input type="file" className="hidden" accept="image/*" onChange={handleImageUpload} />
            </label>
          </div>
          {preview && (
            <div className="mt-4 relative group">
              <img src={preview} alt="Preview" className="w-full h-32 object-cover rounded-xl border border-slate-200" />
              <button 
                type="button"
                onClick={() => { setImage(undefined); setPreview(null); }}
                className="absolute top-2 right-2 bg-red-500 text-white p-1 rounded-full shadow-lg opacity-0 group-hover:opacity-100 transition-opacity"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M6 18L18 6M6 6l12 12" strokeWidth="2" /></svg>
              </button>
            </div>
          )}
        </div>

        <div className="space-y-4 pt-4">
          <label className="text-sm font-semibold text-slate-700 block">Quiz Mode</label>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <button
              type="button"
              onClick={() => setMode(QuizMode.TIMED_QUESTION)}
              className={`p-4 rounded-xl border-2 text-left transition-all ${
                mode === QuizMode.TIMED_QUESTION 
                  ? 'border-indigo-600 bg-indigo-50 ring-2 ring-indigo-200' 
                  : 'border-slate-200 hover:border-slate-300'
              }`}
            >
              <div className="font-bold text-slate-900">Per-Question Timer</div>
              <div className="text-xs text-slate-500 mt-1 text-balance">Intensive focus on accuracy under pressure.</div>
            </button>
            <button
              type="button"
              onClick={() => setMode(QuizMode.OVERALL_TIMED)}
              className={`p-4 rounded-xl border-2 text-left transition-all ${
                mode === QuizMode.OVERALL_TIMED 
                  ? 'border-indigo-600 bg-indigo-50 ring-2 ring-indigo-200' 
                  : 'border-slate-200 hover:border-slate-300'
              }`}
            >
              <div className="font-bold text-slate-900">Total Quiz Timer</div>
              <div className="text-xs text-slate-500 mt-1 text-balance">Comprehensive time management across all concepts.</div>
            </button>
          </div>
        </div>

        <button
          type="submit"
          className="w-full py-4 bg-indigo-600 hover:bg-indigo-700 text-white font-bold rounded-xl shadow-lg shadow-indigo-200 transition-all flex items-center justify-center gap-2 group"
        >
          <span>Launch AI Assessment</span>
          <svg className="w-5 h-5 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        </button>
      </form>
    </div>
  );
};

export default QuizConfig;
