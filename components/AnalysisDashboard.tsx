
import React from 'react';
import { QuizResult, AIAnalysis, QuizQuestion } from '../types';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell } from 'recharts';

interface Props {
  result: QuizResult;
  analysis: AIAnalysis | null;
  questions: QuizQuestion[];
  onReset: () => void;
}

const AnalysisDashboard: React.FC<Props> = ({ result, analysis, questions, onReset }) => {
  const accuracy = (result.correctAnswers / result.totalQuestions) * 100;
  
  const chartData = result.performances.map((p, i) => ({
    name: `Q${i + 1}`,
    time: Math.round(p.timeSpentSeconds),
    correct: p.isCorrect ? 1 : 0
  }));

  return (
    <div className="space-y-8 animate-in fade-in duration-700">
      <div className="flex flex-col sm:flex-row items-center justify-between gap-6 bg-white p-8 rounded-2xl shadow-xl shadow-slate-200/50 border border-slate-100">
        <div>
          <h2 className="text-3xl font-bold text-slate-900">Quiz Completed!</h2>
          <p className="text-slate-500 mt-1">Here is a detailed look at how you performed.</p>
        </div>
        <div className="flex items-center gap-6">
          <div className="text-center">
            <div className="text-4xl font-black text-indigo-600">{Math.round(accuracy)}%</div>
            <div className="text-xs font-bold text-slate-400 uppercase tracking-tighter">Accuracy</div>
          </div>
          <div className="w-px h-12 bg-slate-100"></div>
          <div className="text-center">
            <div className="text-4xl font-black text-slate-800">{result.correctAnswers}/{result.totalQuestions}</div>
            <div className="text-xs font-bold text-slate-400 uppercase tracking-tighter">Score</div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Main Analysis */}
        <div className="lg:col-span-2 space-y-8">
          <section className="bg-white p-6 sm:p-8 rounded-2xl shadow-lg border border-slate-100">
            <h3 className="text-lg font-bold text-slate-900 mb-6 flex items-center gap-2">
              <svg className="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" /></svg>
              AI Pedagogical Insights
            </h3>
            
            {analysis ? (
              <div className="space-y-6">
                <p className="text-slate-600 leading-relaxed italic border-l-4 border-indigo-200 pl-4">
                  "{analysis.summary}"
                </p>
                
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                  <div className="bg-green-50 p-4 rounded-xl">
                    <h4 className="font-bold text-green-800 text-sm mb-3 uppercase flex items-center gap-1">
                      <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20"><path d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" /></svg>
                      Strengths
                    </h4>
                    <ul className="space-y-2">
                      {analysis.strengths.map((s, i) => <li key={i} className="text-green-700 text-sm flex items-start gap-2"><span>•</span> {s}</li>)}
                    </ul>
                  </div>
                  <div className="bg-amber-50 p-4 rounded-xl">
                    <h4 className="font-bold text-amber-800 text-sm mb-3 uppercase flex items-center gap-1">
                      <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20"><path d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" /></svg>
                      Areas to Improve
                    </h4>
                    <ul className="space-y-2">
                      {analysis.weaknesses.map((w, i) => <li key={i} className="text-amber-700 text-sm flex items-start gap-2"><span>•</span> {w}</li>)}
                    </ul>
                  </div>
                </div>

                <div className="bg-indigo-600 p-6 rounded-xl text-white">
                  <h4 className="font-bold mb-4 flex items-center gap-2">
                    <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M9 4.804A7.968 7.968 0 005.5 4c-1.255 0-2.443.29-3.5.804v10A7.969 7.969 0 015.5 14c1.669 0 3.218.51 4.5 1.385A7.962 7.962 0 0114.5 14c1.255 0 2.443.29 3.5.804v-10A7.968 7.968 0 0014.5 4c-1.255 0-2.443.29-3.5.804V12a1 1 0 11-2 0V4.804z" /></svg>
                    Study Plan Recommendations
                  </h4>
                  <ul className="space-y-3">
                    {analysis.recommendations.map((r, i) => (
                      <li key={i} className="bg-white/10 p-3 rounded-lg text-sm flex items-start gap-3">
                        <span className="w-5 h-5 rounded-full bg-white/20 flex items-center justify-center text-xs flex-shrink-0">{i+1}</span>
                        {r}
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-12 text-slate-400 gap-4">
                <div className="w-12 h-12 border-4 border-slate-200 border-t-indigo-500 rounded-full animate-spin"></div>
                <p>Gemini is synthesizing your results...</p>
              </div>
            )}
          </section>

          <section className="bg-white p-6 sm:p-8 rounded-2xl shadow-lg border border-slate-100">
            <h3 className="text-lg font-bold text-slate-900 mb-6 flex items-center justify-between">
              <span>Time Consumption per Question</span>
              <span className="text-xs font-normal text-slate-400">Seconds Spent</span>
            </h3>
            <div className="h-64 w-full">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
                  <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{fill: '#64748b', fontSize: 12}} />
                  <YAxis axisLine={false} tickLine={false} tick={{fill: '#64748b', fontSize: 12}} />
                  <Tooltip 
                    cursor={{fill: '#f8fafc'}}
                    contentStyle={{borderRadius: '12px', border: 'none', boxShadow: '0 10px 15px -3px rgb(0 0 0 / 0.1)'}}
                  />
                  <Bar dataKey="time" radius={[4, 4, 0, 0]}>
                    {chartData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.correct ? '#4f46e5' : '#f43f5e'} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
          </section>
        </div>

        {/* Sidebar Metrics */}
        <div className="space-y-6">
          <div className="bg-white p-6 rounded-2xl shadow-lg border border-slate-100">
            <h4 className="font-bold text-slate-900 mb-4 text-sm uppercase">Quick Stats</h4>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-slate-500 text-sm">Total Time</span>
                <span className="font-bold text-slate-800">{Math.round(result.totalTimeSpent)}s</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-slate-500 text-sm">Avg. Time / Question</span>
                <span className="font-bold text-slate-800">{Math.round(result.totalTimeSpent / result.totalQuestions)}s</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-slate-500 text-sm">Question Revisits</span>
                <span className="font-bold text-slate-800">{result.performances.reduce((acc, p) => acc + p.revisits, 0)}</span>
              </div>
            </div>
          </div>

          <div className="bg-indigo-50 p-6 rounded-2xl border border-indigo-100">
            <h4 className="font-bold text-indigo-900 mb-2 text-sm">Ready for more?</h4>
            <p className="text-indigo-700 text-xs mb-4">You can start a new quiz on a different topic or try the same one again to see if you improve.</p>
            <button
              onClick={onReset}
              className="w-full py-3 bg-indigo-600 text-white rounded-xl font-bold text-sm shadow-md hover:bg-indigo-700 transition-colors"
            >
              Start New Quiz
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AnalysisDashboard;
