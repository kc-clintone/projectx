const app = {
    state: {
        sessionId: localStorage.getItem('ep_session'),
        profile: JSON.parse(localStorage.getItem('ep_profile') || 'null'),
        questions: [],
        currentQuestionIndex: 0,
        userResponses: [],
        performances: [],
        mode: 'TIMED_QUESTION',
        timeLeft: 0,
        timerInterval: null,
        questionStartTime: 0,
        visitCounts: []
    },

    init() {
        if (this.state.sessionId && this.state.profile) {
            this.renderConfig();
            const nameEl = document.getElementById('user-name');
            nameEl.textContent = `Hi, ${this.state.profile.name}`;
            nameEl.classList.add('cursor-pointer', 'hover:text-indigo-600', 'transition-colors');
            nameEl.onclick = () => app.renderProfile();
            document.getElementById('user-info').classList.remove('hidden');
        } else {
            this.renderProfile();
        }
    },

    setLoading(isLoading, text = "Loading...") {
        let el = document.getElementById('loading-overlay');
        
        // Dynamically create overlay if it doesn't exist
        if (!el) {
            el = document.createElement('div');
            el.id = 'loading-overlay';
            el.className = 'hidden fixed inset-0 bg-white/60 backdrop-blur-sm z-[60] flex flex-col items-center justify-center gap-6 text-center px-4';
            el.innerHTML = `
                <div class="relative">
                   <div class="w-20 h-20 border-4 border-indigo-100 rounded-full"></div>
                   <div class="w-20 h-20 border-4 border-indigo-600 border-t-transparent rounded-full animate-spin absolute top-0"></div>
                </div>
                <p id="loading-text" class="text-indigo-900 font-bold animate-pulse text-xl">Processing...</p>
            `;
            document.body.appendChild(el);
        }

        const txt = document.getElementById('loading-text');
        if (isLoading) {
            if (txt) txt.textContent = text;
            el.classList.remove('hidden');
            el.classList.add('fade-in');
        } else {
            el.classList.add('hidden');
            el.classList.remove('fade-in');
        }
    },

    async apiCall(endpoint, body) {
        const headers = { 'Content-Type': 'application/json' };
        if (this.state.sessionId) headers['X-Session-ID'] = this.state.sessionId;

        const res = await fetch(endpoint, {
            method: 'POST',
            headers,
            body: JSON.stringify(body)
        });
        
        // Handle session expiration (happens when Go server restarts)
        if (res.status === 401) {
            this.logout();
            throw new Error("Session expired. Please sign in again.");
        }

        // Handle Rate Limiting (429)
        if (res.status === 429) {
            throw new Error("High traffic (Rate Limit). Please wait 10 seconds and try again.");
        }

        if (!res.ok) throw new Error(await res.text());
        return res.json();
    },

    // --- Views ---

    renderProfile() {
        const p = this.state.profile || {};
        const isEditing = !!this.state.sessionId;
        const container = document.getElementById('app-container');
        container.innerHTML = `
            <div class="min-h-[60vh] flex items-center justify-center">
                <div class="bg-white rounded-3xl shadow-xl shadow-slate-200/50 p-8 sm:p-12 max-w-xl w-full border border-slate-100 fade-in relative overflow-hidden">
                    <div class="absolute top-0 left-0 w-full h-2 bg-gradient-to-r from-indigo-500 to-violet-500"></div>
                    <div class="text-center mb-10">
                        <h2 class="text-3xl font-black text-slate-900 mb-3">${isEditing ? 'Edit Profile' : 'Complete Your Profile'}</h2>
                        <p class="text-slate-500 font-medium">This helps us tailor the difficulty of your assessments.</p>
                    </div>
                    <form onsubmit="app.handleProfileSubmit(event)" class="space-y-8">
                        <div class="space-y-6">
                            <div class="space-y-2">
                                <label class="block text-sm font-bold text-slate-700 ml-1">Full Name</label>
                                <input name="name" value="${p.name || ''}" required class="w-full px-5 py-4 rounded-2xl border border-slate-200 focus:ring-4 focus:ring-indigo-100 focus:border-indigo-500 transition-all outline-none bg-slate-50/50" placeholder="How should we address you?">
                            </div>
                            <div class="space-y-2">
                                <label class="block text-sm font-bold text-slate-700 ml-1">Grade Level / Academic Stage</label>
                                <select name="grade" required class="w-full px-5 py-4 rounded-2xl border border-slate-200 focus:ring-4 focus:ring-indigo-100 focus:border-indigo-500 transition-all outline-none bg-slate-50/50">
                                    <option value="">Select your level</option>
                                    <option value="Elementary" ${p.grade === 'Elementary' ? 'selected' : ''}>Elementary School</option>
                                    <option value="Middle" ${p.grade === 'Middle' ? 'selected' : ''}>Middle School</option>
                                    <option value="High" ${p.grade === 'High' ? 'selected' : ''}>High School</option>
                                    <option value="University" ${p.grade === 'University' ? 'selected' : ''}>University / College</option>
                                    <option value="Professional" ${p.grade === 'Professional' ? 'selected' : ''}>Professional Certification</option>
                                    <option value="Lifelong Learner" ${p.grade === 'Lifelong Learner' ? 'selected' : ''}>Lifelong Learner</option>
                                </select>
                            </div>
                        </div>
                        <div class="flex gap-3">
                            ${isEditing ? `
                            <button type="button" onclick="app.init()" class="flex-1 py-5 bg-slate-100 hover:bg-slate-200 text-slate-600 font-bold rounded-2xl transition-all">
                                Cancel
                            </button>` : ''}
                            <button type="submit" class="flex-[2] py-5 bg-indigo-600 hover:bg-indigo-700 text-white font-bold rounded-2xl shadow-xl shadow-indigo-200 transition-all transform active:scale-[0.98] flex items-center justify-center gap-2">
                                <span>${isEditing ? 'Save Changes' : 'Finish Setup'}</span>
                                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
                                </svg>
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        `;
    },

    renderConfig() {
        const container = document.getElementById('app-container');
        container.innerHTML = `
            <div class="bg-white rounded-2xl shadow-xl shadow-slate-200/50 p-6 sm:p-10 max-w-2xl mx-auto border border-slate-100 fade-in">
                <div class="mb-8">
                    <h2 class="text-2xl font-bold text-slate-900 mb-2">Create New Quiz</h2>
                    <p class="text-slate-500 text-sm">Enter an educational prompt or upload study material to begin.</p>
                    <div class="mt-4 p-3 bg-amber-50 border border-amber-100 rounded-xl flex items-start gap-3">
                        <svg class="w-5 h-5 text-amber-500 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                            <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
                        </svg>
                        <p class="text-[11px] text-amber-800 font-medium leading-relaxed">
                            <strong>Educational Policy:</strong> EduPulse AI is strictly for academic and learning purposes. Topics involving violence, illegal acts, or inappropriate content will be automatically rejected.
                        </p>
                    </div>
                </div>

                <form onsubmit="app.handleConfigSubmit(event)" class="space-y-6">
                    <div class="space-y-2">
                        <label class="text-sm font-semibold text-slate-700 block">Goal</label>
                        <input type="hidden" name="goal" id="input-goal" value="QUIZ">
                        <div class="flex p-1 bg-slate-100 rounded-xl">
                            <button type="button" onclick="app.setGoal('QUIZ')" id="btn-goal-QUIZ" class="flex-1 py-2 rounded-lg text-sm font-bold transition-all bg-white text-indigo-600 shadow-sm">
                                Take a Quiz
                            </button>
                            <button type="button" onclick="app.setGoal('PLAN')" id="btn-goal-PLAN" class="flex-1 py-2 rounded-lg text-sm font-bold transition-all text-slate-500 hover:text-slate-700">
                                Generate Study Plan
                            </button>
                        </div>
                    </div>

                    <div class="space-y-2">
                        <label class="text-sm font-semibold text-slate-700">Educational Prompt</label>
                        <textarea name="topic" required rows="4" class="w-full px-4 py-3 rounded-xl border border-slate-200 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all outline-none resize-none bg-slate-50/30" placeholder="Describe what you want to learn or ask a specific question. e.g., 'Explain the process of photosynthesis and test me on the light-dependent reactions.'"></textarea>
                    </div>

                    <div id="quiz-options" class="space-y-6">
                    <div class="space-y-4">
                        <label class="text-sm font-semibold text-slate-700 block">Question Type</label>
                        <input type="hidden" name="type" id="input-type" value="MCQ">
                        <div class="flex flex-wrap gap-3">
                            <button type="button" onclick="app.selectType('MCQ')" id="btn-type-MCQ" class="flex-1 min-w-[120px] p-3 rounded-xl border-2 transition-all text-center border-indigo-600 bg-indigo-50 ring-2 ring-indigo-100">
                                <div class="font-bold text-sm text-slate-900">Multiple Choice</div>
                                <div class="text-[10px] text-slate-500 leading-none mt-1">Standard assessment</div>
                            </button>
                            <button type="button" onclick="app.selectType('OPEN')" id="btn-type-OPEN" class="flex-1 min-w-[120px] p-3 rounded-xl border-2 transition-all text-center border-slate-100 bg-slate-50/50 hover:border-slate-300">
                                <div class="font-bold text-sm text-slate-900">Open Ended</div>
                                <div class="text-[10px] text-slate-500 leading-none mt-1">Critical thinking</div>
                            </button>
                            <button type="button" onclick="app.selectType('MIXED')" id="btn-type-MIXED" class="flex-1 min-w-[120px] p-3 rounded-xl border-2 transition-all text-center border-slate-100 bg-slate-50/50 hover:border-slate-300">
                                <div class="font-bold text-sm text-slate-900">Combination</div>
                                <div class="text-[10px] text-slate-500 leading-none mt-1">Comprehensive</div>
                            </button>
                        </div>
                    </div>
                    </div>

                    <div class="space-y-2">
                        <label class="text-sm font-semibold text-slate-700">Optional: Study Material</label>
                        <div class="flex items-center justify-center w-full">
                            <label class="flex flex-col items-center justify-center w-full h-24 border-2 border-slate-200 border-dashed rounded-xl cursor-pointer bg-slate-50 hover:bg-slate-100 transition-colors">
                                <div class="flex flex-col items-center justify-center pt-5 pb-6">
                                    <p class="text-xs text-slate-500"><span class="font-bold">Upload image</span> or drag and drop</p>
                                    <p class="text-[10px] text-slate-400">Notes, Diagrams (MAX. 5MB)</p>
                                </div>
                                <input type="file" name="image" class="hidden" accept="image/*" onchange="app.handleImagePreview(this)">
                            </label>
                        </div>
                        <div id="image-preview" class="hidden mt-2 relative group w-24 h-24">
                            <img src="" alt="Preview" class="w-full h-full object-cover rounded-xl border border-slate-200">
                            <button type="button" onclick="app.clearImage()" class="absolute -top-2 -right-2 bg-red-500 text-white p-1 rounded-full shadow-lg">
                                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M6 18L18 6M6 6l12 12" strokeWidth="2" /></svg>
                            </button>
                        </div>
                    </div>

                    <div id="timing-options" class="space-y-4 pt-2">
                        <label class="text-sm font-semibold text-slate-700 block">Timing Strategy</label>
                        <input type="hidden" name="mode" id="input-mode" value="TIMED_QUESTION">
                        <div class="grid grid-cols-2 gap-4">
                            <button type="button" onclick="app.selectMode('TIMED_QUESTION')" id="btn-mode-TIMED_QUESTION" class="p-4 rounded-xl border-2 text-left transition-all border-indigo-600 bg-indigo-50 shadow-sm ring-2 ring-indigo-200">
                                <div class="font-bold text-sm text-slate-900">Per Question</div>
                                <p class="text-[10px] text-slate-500 mt-1 leading-tight">Focus on one challenge at a time. No backtracking or changing answers once submitted.</p>
                            </button>
                            <button type="button" onclick="app.selectMode('OVERALL_TIMED')" id="btn-mode-OVERALL_TIMED" class="p-4 rounded-xl border-2 text-left transition-all border-slate-100 bg-slate-50/50 hover:border-slate-300">
                                <div class="font-bold text-sm text-slate-900">Full Session</div>
                                <p class="text-[10px] text-slate-500 mt-1 leading-tight">Manage your total time across all questions. You can revisit and edit any response.</p>
                            </button>
                        </div>
                    </div>

                    <button type="submit" id="generate-btn" class="w-full py-4 bg-indigo-600 hover:bg-indigo-700 text-white font-bold rounded-xl shadow-lg shadow-indigo-200 transition-all flex items-center justify-center gap-2 group mt-4">
                        <span id="submit-text">Launch AI Assessment</span>
                        <svg class="w-5 h-5 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                        </svg>
                    </button>
                </form>
            </div>
        `;
    },

    setGoal(goal) {
        document.getElementById('input-goal').value = goal;
        const btnQuiz = document.getElementById('btn-goal-QUIZ');
        const btnPlan = document.getElementById('btn-goal-PLAN');
        const quizOpts = document.getElementById('quiz-options');
        const timeOpts = document.getElementById('timing-options');
        const submitTxt = document.getElementById('submit-text');

        if (goal === 'QUIZ') {
            btnQuiz.className = "flex-1 py-2 rounded-lg text-sm font-bold transition-all bg-white text-indigo-600 shadow-sm";
            btnPlan.className = "flex-1 py-2 rounded-lg text-sm font-bold transition-all text-slate-500 hover:text-slate-700";
            quizOpts.classList.remove('hidden');
            timeOpts.classList.remove('hidden');
            submitTxt.textContent = "Launch AI Assessment";
        } else {
            btnPlan.className = "flex-1 py-2 rounded-lg text-sm font-bold transition-all bg-white text-indigo-600 shadow-sm";
            btnQuiz.className = "flex-1 py-2 rounded-lg text-sm font-bold transition-all text-slate-500 hover:text-slate-700";
            quizOpts.classList.add('hidden');
            timeOpts.classList.add('hidden');
            submitTxt.textContent = "Generate Study Plan";
        }
    },

    selectType(type) {
        document.getElementById('input-type').value = type;
        ['MCQ', 'OPEN', 'MIXED'].forEach(t => {
            const btn = document.getElementById(`btn-type-${t}`);
            if (t === type) {
                btn.className = "flex-1 min-w-[120px] p-3 rounded-xl border-2 transition-all text-center border-indigo-600 bg-indigo-50 ring-2 ring-indigo-100";
            } else {
                btn.className = "flex-1 min-w-[120px] p-3 rounded-xl border-2 transition-all text-center border-slate-100 bg-slate-50/50 hover:border-slate-300";
            }
        });
    },

    selectMode(mode) {
        document.getElementById('input-mode').value = mode;
        ['TIMED_QUESTION', 'OVERALL_TIMED'].forEach(m => {
            const btn = document.getElementById(`btn-mode-${m}`);
            if (m === mode) {
                btn.className = "p-4 rounded-xl border-2 text-left transition-all border-indigo-600 bg-indigo-50 shadow-sm ring-2 ring-indigo-200";
            } else {
                btn.className = "p-4 rounded-xl border-2 text-left transition-all border-slate-100 bg-slate-50/50 hover:border-slate-300";
            }
        });
    },

    handleImagePreview(input) {
        if (input.files && input.files[0]) {
            const reader = new FileReader();
            reader.onload = (e) => {
                const previewDiv = document.getElementById('image-preview');
                const img = previewDiv.querySelector('img');
                img.src = e.target.result;
                previewDiv.classList.remove('hidden');
            };
            reader.readAsDataURL(input.files[0]);
        }
    },

    clearImage() {
        const input = document.querySelector('input[name="image"]');
        input.value = '';
        const previewDiv = document.getElementById('image-preview');
        previewDiv.classList.add('hidden');
        previewDiv.querySelector('img').src = '';
    },

    renderQuiz() {
        const q = this.state.questions[this.state.currentQuestionIndex];
        const currentIndex = this.state.currentQuestionIndex;
        const totalQuestions = this.state.questions.length;
        const progress = ((currentIndex + 1) / totalQuestions) * 100;
        const container = document.getElementById('app-container');
        const currentResponse = this.state.userResponses[currentIndex];
        const hasResponse = currentResponse !== null && currentResponse !== undefined && currentResponse !== "";
        
        let inputHtml = '';
        if (q.type === 'MCQ') {
            inputHtml = `<div class="grid grid-cols-1 gap-4">
                ${q.options.map((opt, idx) => {
                    const isSelected = currentResponse === idx.toString();
                    return `
                    <button onclick="app.handleOptionSelect(${idx})" class="group flex items-center text-left p-5 rounded-xl border-2 transition-all ${isSelected ? 'border-indigo-600 bg-indigo-50 shadow-md ring-2 ring-indigo-200' : 'border-slate-100 hover:border-indigo-200 hover:bg-slate-50'}">
                        <div class="w-8 h-8 rounded-full flex items-center justify-center mr-4 transition-colors ${isSelected ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-500'}">
                            ${String.fromCharCode(65 + idx)}
                        </div>
                        <span class="flex-1 font-medium ${isSelected ? 'text-indigo-900' : 'text-slate-600'}">
                            ${opt}
                        </span>
                    </button>
                    `;
                }).join('')}
            </div>`;
        } else {
            inputHtml = `<div class="space-y-4">
                <textarea oninput="app.handleTextInput(this.value)" class="w-full h-48 p-6 rounded-2xl border-2 border-slate-100 focus:border-indigo-500 focus:ring-4 focus:ring-indigo-100 transition-all outline-none resize-none text-slate-700 font-medium leading-relaxed bg-slate-50/50" placeholder="Type your explanation here...">${currentResponse || ''}</textarea>
                <p class="text-xs text-slate-400">EduPulse AI will evaluate your reasoning based on conceptual accuracy.</p>
            </div>`;
        }

        // Navigation Buttons
        let backBtn = '';
        if (this.state.mode === 'OVERALL_TIMED') {
            backBtn = `
            <button onclick="app.prevQuestion()" ${currentIndex === 0 ? 'disabled' : ''} class="flex-1 sm:flex-none flex items-center justify-center gap-2 px-6 py-3 rounded-xl border-2 border-slate-200 text-slate-600 font-bold hover:bg-slate-50 transition-colors disabled:opacity-30">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 19l-7-7 7-7" /></svg>
                Back
            </button>`;
        }

        const nextBtnText = currentIndex < totalQuestions - 1 ? 'Next' : 'Submit Quiz';
        const nextBtnAction = currentIndex < totalQuestions - 1 ? 'app.nextQuestion()' : 'app.finishQuiz()';
        const nextBtnIcon = currentIndex < totalQuestions - 1 ? '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" /></svg>' : '';

        container.innerHTML = `
            <div class="space-y-6 fade-in">
                <!-- Header -->
                <div class="bg-white p-4 rounded-2xl shadow-sm flex items-center justify-between">
                    <div class="flex-1 mr-8">
                        <div class="flex justify-between text-xs font-bold text-slate-400 mb-2 uppercase tracking-widest">
                            <span>Question ${currentIndex + 1} of ${totalQuestions}</span>
                            <span>${Math.round(progress)}%</span>
                        </div>
                        <div class="h-2 bg-slate-100 rounded-full overflow-hidden">
                            <div class="h-full bg-indigo-600 transition-all duration-500 ease-out" style="width: ${progress}%"></div>
                        </div>
                    </div>
                    <div class="flex items-center gap-2 px-4 py-2 rounded-xl border ${this.state.timeLeft < 10 ? 'bg-red-50 border-red-200 text-red-600' : 'bg-slate-50 border-slate-200 text-slate-700'}">
                        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <span class="font-mono font-bold text-lg leading-none">${this.state.timeLeft}s</span>
                    </div>
                </div>

                <!-- Question Card -->
                <div class="bg-white p-6 sm:p-10 rounded-2xl shadow-xl shadow-slate-200/50 border border-slate-100 min-h-[400px]">
                    <div class="flex items-center gap-2 mb-6">
                        <span class="px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-widest ${q.type === 'MCQ' ? 'bg-indigo-100 text-indigo-700' : 'bg-violet-100 text-violet-700'}">
                            ${q.type === 'MCQ' ? 'Choice' : 'Explanation'}
                        </span>
                    </div>
                    
                    <h3 class="text-xl sm:text-2xl font-semibold text-slate-800 mb-8 leading-relaxed">
                        ${q.text}
                    </h3>

                    ${inputHtml}
                </div>

                <!-- Footer -->
                <div class="flex flex-col sm:flex-row items-center gap-4">
                    <div class="flex items-center gap-3 w-full sm:w-auto">
                        ${backBtn}
                        <button onclick="app.finishEarly()" class="flex-1 sm:flex-none px-6 py-3 rounded-xl border-2 border-slate-100 text-slate-400 font-bold hover:border-red-200 hover:text-red-600 hover:bg-red-50 transition-all">
                            Finish Early
                        </button>
                    </div>
                    
                    <div class="w-full sm:w-auto sm:ml-auto">
                        <button id="quiz-next-btn" onclick="${nextBtnAction}" ${!hasResponse ? 'disabled' : ''} class="w-full sm:w-auto flex items-center justify-center gap-2 px-8 py-3 rounded-xl bg-indigo-600 text-white font-bold hover:bg-indigo-700 shadow-lg shadow-indigo-200 transition-all disabled:opacity-50">
                            ${nextBtnText}
                            ${nextBtnIcon}
                        </button>
                    </div>
                </div>
            </div>
        `;
    },

    renderStudyPlan(plan) {
        const container = document.getElementById('app-container');
        container.innerHTML = `
            <div class="space-y-8 fade-in max-w-4xl mx-auto">
                <div class="bg-white p-8 rounded-2xl shadow-xl shadow-slate-200/50 border border-slate-100">
                    <div class="flex items-start justify-between gap-4 mb-6">
                        <div>
                            <span class="text-xs font-bold text-indigo-600 uppercase tracking-widest">Personalized Study Plan</span>
                            <h2 class="text-3xl font-black text-slate-900 mt-2">${plan.title}</h2>
                        </div>
                        <button onclick="app.renderConfig()" class="text-slate-400 hover:text-slate-600 transition-colors">
                            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" /></svg>
                        </button>
                    </div>
                    <p class="text-slate-600 leading-relaxed text-lg border-l-4 border-indigo-200 pl-4 mb-8">${plan.introduction}</p>
                    
                    <div class="space-y-6">
                        ${plan.modules.map((mod, idx) => `
                            <div class="bg-slate-50 rounded-xl p-6 border border-slate-100 hover:border-indigo-200 transition-colors group">
                                <div class="flex items-center justify-between mb-3">
                                    <h3 class="text-lg font-bold text-slate-800 flex items-center gap-3">
                                        <span class="w-8 h-8 rounded-lg bg-indigo-600 text-white flex items-center justify-center text-sm font-bold shadow-md shadow-indigo-200">${idx + 1}</span>
                                        ${mod.topic}
                                    </h3>
                                    <span class="text-xs font-bold px-3 py-1 bg-white rounded-full text-slate-500 border border-slate-200 flex items-center gap-1">
                                        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                                        ${mod.timeEstimate}
                                    </span>
                                </div>
                                <p class="text-slate-600 mb-4 ml-11">${mod.description}</p>
                                <div class="ml-11 bg-white p-4 rounded-lg border border-slate-200 text-sm text-slate-700 flex items-start gap-2">
                                    <svg class="w-5 h-5 text-indigo-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" /></svg>
                                    <span><strong>Activity:</strong> ${mod.activity}</span>
                                </div>
                            </div>
                        `).join('')}
                    </div>
                </div>
            </div>
        `;
    },

    renderAnalysis(analysis, result) {
        const container = document.getElementById('app-container');
        const gradedPerformances = analysis.gradedPerformances || [];
        const finalCorrectCount = gradedPerformances.length > 0 
            ? gradedPerformances.filter(gp => gp.isCorrect).length 
            : result.correctAnswers;
        
        const accuracy = (finalCorrectCount / result.totalQuestions) * 100;

        // Generate Chart Data (CSS Bar Chart)
        const maxTime = Math.max(...result.performances.map(p => p.timeSpentSeconds), 1);
        const chartBars = result.performances.map((p, i) => {
            const gp = gradedPerformances.find(g => g.questionId === p.questionId);
            const isCorrect = gp ? gp.isCorrect : p.isCorrect;
            const height = Math.max((p.timeSpentSeconds / maxTime) * 100, 5); // Min 5% height
            const colorClass = isCorrect ? 'bg-indigo-600' : 'bg-rose-500';
            return `
                <div class="flex flex-col items-center gap-2 group w-full h-full justify-end">
                    <div class="relative w-full bg-slate-100 rounded-t-md h-full flex items-end justify-center overflow-hidden">
                        <div class="${colorClass} w-full mx-1 rounded-t-sm transition-all duration-500" style="height: ${height}%"></div>
                        <div class="absolute bottom-0 mb-2 opacity-0 group-hover:opacity-100 transition-opacity bg-slate-800 text-white text-[10px] p-1 rounded pointer-events-none z-10">
                            ${Math.round(p.timeSpentSeconds)}s
                        </div>
                    </div>
                    <span class="text-xs text-slate-400 font-medium">Q${i+1}</span>
                </div>
            `;
        }).join('');
        
        container.innerHTML = `
            <div class="space-y-8 fade-in">
                <div class="flex flex-col sm:flex-row items-center justify-between gap-6 bg-white p-8 rounded-2xl shadow-xl shadow-slate-200/50 border border-slate-100">
                    <div>
                        <h2 class="text-3xl font-bold text-slate-900">Performance Report</h2>
                        <p class="text-slate-500 mt-1">Comprehensive evaluation complete.</p>
                    </div>
                    <div class="flex items-center gap-6">
                        <div class="text-center">
                            <div class="text-4xl font-black text-indigo-600">${Math.round(accuracy)}%</div>
                            <div class="text-xs font-bold text-slate-400 uppercase tracking-tighter">Accuracy</div>
                        </div>
                        <div class="w-px h-12 bg-slate-100"></div>
                        <div class="text-center">
                            <div class="text-4xl font-black text-slate-800">${finalCorrectCount}/${result.totalQuestions}</div>
                            <div class="text-xs font-bold text-slate-400 uppercase tracking-tighter">Final Score</div>
                        </div>
                    </div>
                </div>

                <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
                    <div class="lg:col-span-2 space-y-8">
                        <section class="bg-white p-6 sm:p-8 rounded-2xl shadow-lg border border-slate-100">
                            <h3 class="text-lg font-bold text-slate-900 mb-6 flex items-center gap-2">
                                <svg class="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" /></svg>
                                AI Pedagogical Analysis
                            </h3>
                            
                            <div class="space-y-6">
                                <p class="text-slate-600 leading-relaxed italic border-l-4 border-indigo-200 pl-4">"${analysis.summary}"</p>
                                
                                <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
                                    <div class="bg-green-50 p-4 rounded-xl border border-green-100">
                                        <h4 class="font-bold text-green-800 text-xs mb-3 uppercase flex items-center gap-1 tracking-wider">Strengths</h4>
                                        <ul class="space-y-2">
                                            ${analysis.strengths.map(s => `<li class="text-green-700 text-sm flex items-start gap-2"><span>•</span> ${s}</li>`).join('')}
                                        </ul>
                                    </div>
                                    <div class="bg-amber-50 p-4 rounded-xl border border-amber-100">
                                        <h4 class="font-bold text-amber-800 text-xs mb-3 uppercase flex items-center gap-1 tracking-wider">Growth Areas</h4>
                                        <ul class="space-y-2">
                                            ${analysis.weaknesses.map(w => `<li class="text-amber-700 text-sm flex items-start gap-2"><span>•</span> ${w}</li>`).join('')}
                                        </ul>
                                    </div>
                                </div>

                                <div class="space-y-4">
                                    <h4 class="text-sm font-bold text-slate-400 uppercase tracking-widest px-1">Detailed Question Feedback</h4>
                                    <div class="space-y-3">
                                        ${result.performances.map((p, i) => {
                                            const grade = gradedPerformances.find(gp => gp.questionId === p.questionId);
                                            const question = this.state.questions[i];
                                            const isMCQ = question.type === 'MCQ';
                                            const isCorrect = grade ? grade.isCorrect : p.isCorrect;
                                            
                                            return `
                                            <div class="p-4 rounded-xl border-2 transition-all ${isCorrect ? 'bg-green-50/30 border-green-100' : 'bg-red-50/30 border-red-100'}">
                                                <div class="flex items-center justify-between mb-2">
                                                    <span class="text-[10px] font-black uppercase tracking-widest text-slate-400">Question ${i + 1} (${question.type})</span>
                                                    <span class="text-[10px] font-bold px-2 py-0.5 rounded-full ${isCorrect ? 'bg-green-500 text-white' : 'bg-red-500 text-white'}">
                                                        ${isCorrect ? 'Correct' : 'Incorrect'}
                                                    </span>
                                                </div>
                                                <p class="text-sm font-semibold text-slate-800 mb-2">${question.text}</p>
                                                ${!isMCQ ? `
                                                <div class="mb-3 p-3 bg-white/60 rounded-lg text-xs italic text-slate-600 border border-slate-100">
                                                    <span class="font-bold text-slate-400 not-italic mr-1">Your Response:</span>
                                                    ${p.userResponse || 'No answer provided.'}
                                                </div>` : ''}
                                                <p class="text-xs text-slate-600 leading-relaxed">
                                                    <span class="font-bold text-slate-500">AI Feedback:</span> ${grade ? grade.aiFeedback : 'No feedback available.'}
                                                </p>
                                            </div>`;
                                        }).join('')}
                                    </div>
                                </div>

                                <div class="bg-indigo-600 p-6 rounded-xl text-white shadow-xl shadow-indigo-100">
                                    <h4 class="font-bold mb-4 flex items-center gap-2">
                                        <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M9 4.804A7.968 7.968 0 005.5 4c-1.255 0-2.443.29-3.5.804v10A7.969 7.969 0 015.5 14c1.669 0 3.218.51 4.5 1.385A7.962 7.962 0 0114.5 14c1.255 0 2.443.29 3.5.804v-10A7.968 7.968 0 0014.5 4c-1.255 0-2.443.29-3.5.804V12a1 1 0 11-2 0V4.804z" /></svg>
                                        Personalized Study Plan
                                    </h4>
                                    <ul class="space-y-3">
                                        ${analysis.recommendations.map((r, i) => `
                                            <li class="bg-white/10 p-3 rounded-lg text-sm flex items-start gap-3 border border-white/5">
                                                <span class="w-5 h-5 rounded-full bg-white/20 flex items-center justify-center text-xs flex-shrink-0 font-bold">${i+1}</span>
                                                ${r}
                                            </li>
                                        `).join('')}
                                    </ul>
                                </div>
                            </div>
                        </section>

                        <section class="bg-white p-6 sm:p-8 rounded-2xl shadow-lg border border-slate-100">
                            <h3 class="text-lg font-bold text-slate-900 mb-6 flex items-center justify-between">
                                <span>Engagement Metrics</span>
                                <span class="text-xs font-normal text-slate-400">Seconds Spent</span>
                            </h3>
                            <div class="h-64 w-full flex items-end justify-between gap-2 px-4">
                                ${chartBars}
                            </div>
                        </section>
                    </div>
                    
                    <div class="space-y-6">
                        <div class="bg-white p-6 rounded-2xl shadow-lg border border-slate-100">
                            <h4 class="font-bold text-slate-900 mb-4 text-xs uppercase tracking-widest">Efficiency Stats</h4>
                            <div class="space-y-4">
                                <div class="flex items-center justify-between">
                                    <span class="text-slate-500 text-sm">Total Session</span>
                                    <span class="font-bold text-slate-800">${Math.round(result.totalTimeSpent)}s</span>
                                </div>
                                <div class="flex items-center justify-between">
                                    <span class="text-slate-500 text-sm">Avg. Pace</span>
                                    <span class="font-bold text-slate-800">${Math.round(result.totalTimeSpent / result.totalQuestions)}s/q</span>
                                </div>
                                <div class="flex items-center justify-between">
                                    <span class="text-slate-500 text-sm">Concept Revisits</span>
                                    <span class="font-bold text-slate-800">${result.performances.reduce((acc, p) => acc + p.revisits, 0)}</span>
                                </div>
                            </div>
                        </div>

                        <div class="bg-slate-900 p-6 rounded-2xl text-white">
                            <h4 class="font-bold mb-2 text-sm">Continue Learning</h4>
                            <p class="text-slate-400 text-xs mb-4 leading-relaxed">Your results are saved to your pedagogical history. Ready for the next topic?</p>
                            <button onclick="app.renderConfig()" class="w-full py-3 bg-indigo-500 text-white rounded-xl font-bold text-sm shadow-lg hover:bg-indigo-400 transition-colors">
                                New Session
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        `;
    },

    // --- Handlers ---

    async handleProfileSubmit(e) {
        e.preventDefault();
        const formData = new FormData(e.target);
        const profile = { name: formData.get('name'), grade: formData.get('grade') };
        
        // If session exists, update profile instead of creating new session
        if (this.state.sessionId) {
            this.setLoading(true, "Updating Profile...");
            try {
                await this.apiCall('/api/profile', profile);
                this.state.profile = profile;
                localStorage.setItem('ep_profile', JSON.stringify(profile));
                this.init(); // Return to dashboard
            } catch (err) {
                alert("Update failed: " + err.message);
            } finally {
                this.setLoading(false);
            }
            return;
        }

        this.setLoading(true, "Creating Session...");
        try {
            const res = await this.apiCall('/api/login', profile);
            this.state.sessionId = res.sessionId;
            this.state.profile = profile;
            
            localStorage.setItem('ep_session', res.sessionId);
            localStorage.setItem('ep_profile', JSON.stringify(profile));
            
            this.init();
        } catch (err) {
            alert("Login failed: " + err.message);
        } finally {
            this.setLoading(false);
        }
    },

    async handleConfigSubmit(e) {
        e.preventDefault();
        this.setLoading(true, "AI is crafting your assessment...");
        const formData = new FormData(e.target);
        
        // Handle Image Upload
        let imageBase64 = "";
        const file = formData.get('image');
        if (file && file.size > 0) {
            try {
                imageBase64 = await new Promise((resolve, reject) => {
                    const reader = new FileReader();
                    reader.onload = () => resolve(reader.result);
                    reader.onerror = reject;
                    reader.readAsDataURL(file);
                });
            } catch (err) {
                console.error("Image read error", err);
                alert("Failed to read image file.");
                this.setLoading(false);
                return;
            }
        }

        const req = {
            topic: formData.get('topic'),
            goal: formData.get('goal'),
            mode: formData.get('mode'),
            type: formData.get('type'),
            image: imageBase64
        };

        try {
            const res = await this.apiCall('/api/generate', req);
            if (!res.isValid) {
                alert("Topic rejected: " + res.reason);
                return;
            }
            
            if (res.studyPlan) {
                this.renderStudyPlan(res.studyPlan);
                return;
            }

            this.state.questions = res.questions;
            this.state.mode = req.mode;
            this.state.userResponses = new Array(res.questions.length).fill(null);
            this.state.performances = res.questions.map(q => ({
                questionId: q.id,
                timeSpentSeconds: 0,
                attempts: 0,
                isCorrect: false,
                revisits: 0,
                userResponse: ""
            }));
            this.state.visitCounts = new Array(res.questions.length).fill(0);
            this.state.currentQuestionIndex = 0;
            
            this.startQuiz();
        } catch (err) {
            alert("Generation failed: " + err.message);
        } finally {
            this.setLoading(false);
        }
    },

    // ... (Rest of the methods: startQuiz, startTimer, recordPerformance, handleOptionSelect, handleTextInput, nextQuestion, finishQuiz, logout)
    // Re-implementing standard methods to ensure full file integrity

    startQuiz() {
        this.state.visitCounts[0] = 1;
        this.state.questionStartTime = Date.now();
        
        if (this.state.mode === 'TIMED_QUESTION') {
            this.state.timeLeft = this.state.questions[0].timeLimitSeconds;
        } else {
            this.state.timeLeft = this.state.questions.reduce((acc, q) => acc + q.timeLimitSeconds, 0);
        }

        this.startTimer();
        this.renderQuiz();
    },

    startTimer() {
        if (this.state.timerInterval) clearInterval(this.state.timerInterval);
        this.state.timerInterval = setInterval(() => {
            this.state.timeLeft--;
            
            // Update timer UI directly
            const timerEl = document.querySelector('.font-mono');
            if (timerEl) {
                timerEl.textContent = this.state.timeLeft + 's';
                if (this.state.timeLeft < 10) timerEl.classList.add('text-red-600');
            }

            if (this.state.timeLeft <= 0) {
                if (this.state.mode === 'TIMED_QUESTION') {
                    this.nextQuestion();
                } else {
                    this.finishQuiz();
                }
            }
        }, 1000);
    },

    recordPerformance() {
        const now = Date.now();
        const delta = (now - this.state.questionStartTime) / 1000;
        this.state.questionStartTime = now;
        
        const idx = this.state.currentQuestionIndex;
        const p = this.state.performances[idx];
        p.timeSpentSeconds += delta;
        
        const resp = this.state.userResponses[idx];
        const q = this.state.questions[idx];
        if (q.type === 'MCQ' && resp && q.options[resp]) {
            p.userResponse = q.options[resp];
        } else {
            p.userResponse = resp || "";
        }
        
        p.revisits = (this.state.visitCounts[idx] || 1) - 1;
    },

    handleOptionSelect(idx) {
        const currentIdx = this.state.currentQuestionIndex;
        this.state.userResponses[currentIdx] = idx.toString();
        
        // Update performance immediately for MCQ logic
        const p = this.state.performances[currentIdx];
        p.attempts = (p.attempts || 0) + 1;
        p.isCorrect = idx == this.state.questions[currentIdx].correctIndex;
        p.userResponse = this.state.questions[currentIdx].options[idx];

        this.renderQuiz(); // Re-render to show selection
    },

    handleTextInput(val) {
        this.state.userResponses[this.state.currentQuestionIndex] = val;
        const btn = document.getElementById('quiz-next-btn');
        if (btn) btn.disabled = !val;
    },

    nextQuestion() {
        this.recordPerformance();
        if (this.state.currentQuestionIndex < this.state.questions.length - 1) {
            this.state.currentQuestionIndex++;
            this.state.visitCounts[this.state.currentQuestionIndex] = (this.state.visitCounts[this.state.currentQuestionIndex] || 0) + 1;
            
            if (this.state.mode === 'TIMED_QUESTION') {
                this.state.timeLeft = this.state.questions[this.state.currentQuestionIndex].timeLimitSeconds;
            }
            this.renderQuiz();
        } else {
            this.finishQuiz();
        }
    },

    prevQuestion() {
        if (this.state.mode === 'OVERALL_TIMED' && this.state.currentQuestionIndex > 0) {
            this.recordPerformance();
            this.state.currentQuestionIndex--;
            this.state.visitCounts[this.state.currentQuestionIndex] = (this.state.visitCounts[this.state.currentQuestionIndex] || 0) + 1;
            this.renderQuiz();
        }
    },

    finishEarly() {
        if (confirm("Are you sure you want to end the quiz early? Your progress will be submitted for analysis.")) {
            this.finishQuiz();
        }
    },

    async finishQuiz() {
        clearInterval(this.state.timerInterval);
        this.recordPerformance();
        
        const result = {
            totalQuestions: this.state.questions.length,
            correctAnswers: 0, // Calculated by backend
            totalTimeSpent: this.state.performances.reduce((acc, p) => acc + p.timeSpentSeconds, 0),
            performances: this.state.performances,
            mode: this.state.mode
        };

        this.setLoading(true, "Evaluating your performance...");
        try {
            const analysis = await this.apiCall('/api/analyze', result);
            this.renderAnalysis(analysis, result);
        } catch (err) {
            alert("Analysis failed: " + err.message);
        } finally {
            this.setLoading(false);
        }
    },

    logout() {
        localStorage.removeItem('ep_session');
        localStorage.removeItem('ep_profile');
        location.reload();
    }
};

// Start App
document.addEventListener('DOMContentLoaded', () => app.init());
