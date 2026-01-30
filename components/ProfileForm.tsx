
import React, { useState } from 'react';
import { UserProfile } from '../types';

interface Props {
  onSubmit: (data: UserProfile) => void;
  initialData: UserProfile | null;
}

const ProfileForm: React.FC<Props> = ({ onSubmit, initialData }) => {
  const [formData, setFormData] = useState<UserProfile>(initialData || {
    name: '',
    grade: ''
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name || !formData.grade) return;
    onSubmit(formData);
  };

  return (
    <div className="bg-white rounded-3xl shadow-xl shadow-slate-200/50 p-8 sm:p-12 max-w-xl mx-auto border border-slate-100">
      <div className="text-center mb-10">
        <h2 className="text-3xl font-black text-slate-900 mb-3">Complete Your Profile</h2>
        <p className="text-slate-500 font-medium">This helps us tailor the difficulty of your assessments.</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-8">
        <div className="space-y-6">
          <div className="space-y-2">
            <label className="text-sm font-bold text-slate-700 ml-1">Full Name</label>
            <input
              required
              className="w-full px-5 py-4 rounded-2xl border border-slate-200 focus:ring-4 focus:ring-indigo-100 focus:border-indigo-500 transition-all outline-none bg-slate-50/50"
              placeholder="How should we address you?"
              value={formData.name}
              onChange={e => setFormData({...formData, name: e.target.value})}
            />
          </div>
          
          <div className="space-y-2">
            <label className="text-sm font-bold text-slate-700 ml-1">Grade Level / Academic Stage</label>
            <select
              required
              className="w-full px-5 py-4 rounded-2xl border border-slate-200 focus:ring-4 focus:ring-indigo-100 focus:border-indigo-500 transition-all outline-none bg-slate-50/50 appearance-none"
              value={formData.grade}
              onChange={e => setFormData({...formData, grade: e.target.value})}
            >
              <option value="">Select your level</option>
              <option value="Elementary">Elementary School</option>
              <option value="Middle">Middle School</option>
              <option value="High">High School</option>
              <option value="University">University / College</option>
              <option value="Professional">Professional Certification</option>
              <option value="Lifelong Learner">Lifelong Learner</option>
            </select>
          </div>
        </div>

        <button
          type="submit"
          className="w-full py-5 bg-indigo-600 hover:bg-indigo-700 text-white font-bold rounded-2xl shadow-xl shadow-indigo-200 transition-all transform active:scale-[0.98] flex items-center justify-center gap-2"
        >
          <span>Finish Setup</span>
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
          </svg>
        </button>
      </form>
    </div>
  );
};

export default ProfileForm;
