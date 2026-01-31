// Use global UMD build loaded in index.html to avoid bare module specifier issues
const { h, render } = window.preact;
const { useState, useEffect, useRef } = window.preactHooks;

const api = {
  register: (body)=>fetch('/api/v1/register',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify(body)}).then(r=>({ok:r.ok,status:r.status})),
  login: (body)=>fetch('/api/v1/login',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify(body)}).then(r=>({ok:r.ok,status:r.status})),
  me: ()=>fetch('/api/v1/me').then(r=>r.ok?r.json():null),
  createSession: (body)=>fetch('/api/v1/session',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify(body)}).then(r=>r.json()),
  submitResults: (body)=>fetch('/api/v1/submit_results',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify(body)}).then(r=>r.json()),
  profile: (id)=>fetch(`/api/v1/student/${id}/profile`).then(r=>r.ok?r.json():null)
}

// ConfettiCanvas: lightweight celebratory animation used after achievements
function ConfettiCanvas(){
  const ref = useRef(null);
  useEffect(()=>{
    const canvas = ref.current; if(!canvas) return;
    const ctx = canvas.getContext('2d');
    let running = true;
    function resize(){ canvas.width = window.innerWidth; canvas.height = window.innerHeight }
    resize(); window.addEventListener('resize', resize);
    const particles = [];
    function emit(){ for(let i=0;i<60;i++){ particles.push({x:Math.random()*canvas.width,y:-10, vx:(Math.random()-0.5)*6, vy:Math.random()*4+2, life:Math.random()*60+60, color:['#f59e0b','#ef4444','#10b981','#3b82f6'][Math.floor(Math.random()*4)]}) } }
    function tick(){ ctx.clearRect(0,0,canvas.width,canvas.height); for(let i=particles.length-1;i>=0;i--){ const p=particles[i]; p.x+=p.vx; p.y+=p.vy; p.vy+=0.15; p.life--; ctx.fillStyle=p.color; ctx.fillRect(p.x,p.y,6,8); if(p.life<=0||p.y>canvas.height+20) particles.splice(i,1) } if(running) requestAnimationFrame(tick) }
    emit(); tick(); const t = setTimeout(()=>{running=false; window.removeEventListener('resize', resize)},4000);
    return ()=>{ running=false; clearTimeout(t); window.removeEventListener('resize', resize) }
  },[]);
  return h('canvas',{ref:ref,class:'confetti-canvas'});
}

function Button(props){
  const disabled = props.disabled ? {opacity:0.6,cursor:'not-allowed'} : {}
  return h('button',Object.assign({class:'btn',style:disabled,disabled:props.disabled, onClick:props.onClick},{}),props.children)
}

function Login({onLogin}){
  const [username,setUsername]=useState('')
  const [password,setPassword]=useState('')
  const [email,setEmail]=useState('')
  const [mode,setMode]=useState('login')
  const [error,setError]=useState('')
  const [loading,setLoading]=useState(false)

  async function submit(){
    setError('')
    if(!username) { setError('Please enter a username'); return }
    if(!password) { setError('Please enter a password'); return }
    setLoading(true)
    try{
      if(mode==='register'){
        const r = await api.register({username,password,email})
        if(r.ok) { alert('Registered, please login'); setMode('login'); setPassword('') }
        else setError('Registration failed (user may already exist)')
      } else {
        const r = await api.login({username,password})
        if(r.ok) { localStorage.setItem('sc_username', username); onLogin(username) }
        else setError('Login failed: invalid credentials')
      }
    }catch(e){ setError('Network error') }
    setLoading(false)
  }

  return h('div',{class:'card'},
    h('h2',null,'Welcome to Study Coach'),
    error ? h('div',{style:{color:'crimson',marginBottom:8}}, error) : null,
    h('div',null, h('label',null,'Username'), h('input',{value:username,onInput:e=>setUsername(e.target.value)})),
    h('div',null, h('label',null,'Password'), h('input',{type:'password',value:password,onInput:e=>setPassword(e.target.value)})),
    mode==='register' ? h('div',null, h('label',null,'Email'), h('input',{type:'email',value:email,onInput:e=>setEmail(e.target.value)})) : null,
    h('div',{style:{marginTop:8}},
      h('button',{onClick:()=>setMode(mode==='login'?'register':'login')}, mode==='login'?'Register':'Back to Login'), ' ',
      h(Button,{onClick:submit,disabled:loading}, mode==='login'?(loading?'Logging in...':'Login'):(loading?'Registering...':'Create Account'))
    )
  )
}

function formatDate(ts){ try{ const d = new Date(ts); return d.toLocaleString() }catch(e){ return ts }
}

// update Dashboard rendering to include timeline if present
function Dashboard({onStart, onNewAchievements}){
  const [me,setMe]=useState(null)
  const [error,setError]=useState('')
  useEffect(()=>{
    api.me().then(u=>{
      if(u) {
        setMe(u)
        const achievements = (u.profile && u.profile.achievements) || u.achievements || []
        const key = 'sc_achievements_'+(u.username||'guest')
        const storedRaw = localStorage.getItem(key)
        if(storedRaw === null) {
          try{ localStorage.setItem(key, JSON.stringify(achievements || [])) } catch(e){}
        } else {
          try{
            const stored = JSON.parse(storedRaw || '[]')
            const newOnes = (achievements || []).filter(a=>stored.indexOf(a) === -1)
            if(newOnes.length>0) {
              if(typeof onNewAchievements === 'function') onNewAchievements(newOnes)
              localStorage.setItem(key, JSON.stringify(Array.from(new Set([].concat(stored, achievements || [])))))
            }
          } catch(e) { }
        }
      } else {
        setError('Not authenticated')
      }
    }).catch(()=>setError('Failed to load profile'))
  },[])

  // helper renderers
  function Avatar({name}){
    const initials = (name||'').split(' ').map(s=>s[0]||'').join('').slice(0,2).toUpperCase()
    return h('div',{class:'avatar'}, initials)
  }

  function Stat({label,value}){
    return h('div',{class:'stat'}, h('div',{class:'stat-value'}, value), h('div',{class:'stat-label'}, label))
  }

  if(!me) return h('div',{class:'card'}, error || 'Loading...')

  const profile = me.profile || {}
  const achievements = profile.achievements || []
  const badges = profile.earned_badges || []
  const recent = profile.recent_achievements || []

  return h('div',{class:'dashboard app'},
    h('div',{class:'left column'},
      h('div',{class:'profile-card card'},
        h(Avatar,{name:me.username}),
        h('h2',null, me.username),
        h('div',{class:'muted'}, me.email || ''),
        h('div',{style:{height:8}}),
        h('div',{class:'stats-row'}, h(Stat,{label:'Points',value:profile.points||0}), h(Stat,{label:'Streak',value:profile.streak||0}), h(Stat,{label:'Badges',value:badges.length||0})),
        h('div',{style:{height:8}}),
        h('div',null, h('h4',null,'Badges')),
        h('div',{class:'badges-row'}, badges.map(b=> h('div',{class:'badge-chip',style:{background:b.color}}, h('span',{class:'icon'}, b.icon), h('span',null,b.name)))),
        h('div',{style:{height:12}}),
        h('div',null, h('h4',null,'Achievements')),
        h('ul',{class:'achievements-list'}, achievements.map(a=> h('li',null,a))),
        h('div',{style:{height:12}}),
        h(Button,{onClick:onStart}, 'Start New Session')
      )
    ),
    h('div',{class:'right column'},
      h('div',{class:'card'}, h('h3',null,'Recent Activities'),
        recent.length ? h('ul',{class:'activity-list'}, recent.map(r=> h('li',null, h('div',{class:'activity-title'}, r.name), h('div',{class:'activity-time muted'}, formatDate(r.earned_at))))) : h('div',null,'No recent activity')
      ),
      h('div',{style:{height:12}}),
      h('div',{class:'card'}, h('h3',null,'Latest Study Plan'), profile && profile.last_plan ? h('div',null, h('div',null,'Subject: '+(profile.last_plan.subject||'—')), h('div',null, h('h4',null,'Focus:')), h('ul',null, Object.keys(profile.last_plan.focus||{}).map(k=> h('li',null, h('strong',null,k), ': ', profile.last_plan.focus[k])))) : h('div',null,'No study plan yet')
      )
    )
  )
}

function CreateSession({onCreated}){
  const [tasksText,setTasksText]=useState('')
  const [subject,setSubject]=useState('math')
  const [image,setImage]=useState(null)
  const [ocrRunning,setOcrRunning]=useState(false)
  const [error,setError]=useState('')

  function onFile(e){ const f=e.target.files[0]; if(!f) return; setError(''); if(f.type.startsWith('image/')) { setImage(f); runOcr(f) } else { const r=new FileReader(); r.onload=()=>setTasksText(r.result.split('\n').slice(0,20).join('\n')); r.readAsText(f)} }

  async function runOcr(file){ setOcrRunning(true); setError(''); try{ const { createWorker } = Tesseract; const worker = createWorker({logger:m=>console.log(m)}); await worker.load(); await worker.loadLanguage('eng'); await worker.initialize('eng'); const { data } = await worker.recognize(file); await worker.terminate(); setTasksText(data.text || ''); } catch(e){ setError('OCR failed'); } finally{ setOcrRunning(false) } }

  async function submit(){ setError(''); if(!subject) { setError('Please enter a subject'); return } const tasks = tasksText.split('\n').filter(Boolean).map((t,i)=>({id:String(i+1),prompt:t})); if(tasks.length===0){ setError('Please provide at least one task'); return } try{ const sess = await api.createSession({student_id:localStorage.getItem('sc_username')||'demo',student_level:'grade10',subject,tasks}); onCreated(sess) } catch(e){ setError('Failed to create session') } }

  return h('div',{class:'card'},
    h('h2',null,'Create Session'),
    error ? h('div',{style:{color:'crimson'}},error) : null,
    h('div',null,'Subject: ', h('input',{value:subject,onInput:e=>setSubject(e.target.value)})),
    h('div',null,'Paste or upload tasks:'),
    h('textarea',{style:{width:'100%',height:140},value:tasksText,onInput:(e)=>setTasksText(e.target.value)}),
    h('div',null, h('input',{type:'file',onChange:onFile}), ocrRunning? h('div',null,'OCR running...'):null),
    h('div',{style:{marginTop:8}}, h(Button,{onClick:submit},'Create Session'))
  )
}

function SessionPlayer({session,onDone}){
  const [idx,setIdx]=useState(0)
  const [elapsed,setElapsed]=useState(0)
  const [running,setRunning]=useState(false)
  useEffect(()=>{ let t; if(running) t=setInterval(()=>setElapsed(e=>e+1),1000); return ()=>clearInterval(t) },[running])
  function start(){ setElapsed(0); setRunning(true) }
  function next(){ if(!session) return; session.tasks[idx].actual_seconds = elapsed; setElapsed(0); if(idx+1>=session.tasks.length){ setRunning(false); onDone(session) } else { setIdx(idx+1) } }
  const current = session.tasks[idx] || {estimated_secs:0,prompt:'(none)'}
  const remaining = Math.max(0, (current.estimated_secs || 0) - elapsed)
  return h('div',{class:'card'}, h('h2',null,'Session Player'), h('div',null,'Task: ', current.prompt), h('div',{class:'taskLarge'}, 'Remaining: '+remaining+'s'), h('div',{style:{marginTop:8}}, h(Button,{onClick:start,disabled:running||remaining===0},'Start'), ' ', h(Button,{onClick:next},'Done/Next')) )
}

function StudyPlan({plan, onClose}){
  if(!plan) return null;
  return h('div',{class:'card'},
    h('h2',null,'Study Plan'),
    h('div',null, plan.subject ? h('div',null,'Subject: '+plan.subject) : null),
    h('div',null, h('h3',null,'Focus areas:')),
    h('ul',null, Object.keys(plan.focus || {}).map(k=> h('li',null, h('strong',null,k), ': ', plan.focus[k]))),
    plan.next_timers && plan.next_timers.length ? h('div',null, h('h3',null,'Suggested timers (per task):'), h('ol',null, plan.next_timers.map((t,i)=> h('li',null, 'Task '+(i+1)+': '+t+'s')))) : null,
    h('div',{style:{marginTop:8}}, h(Button,{onClick:onClose},'Back to Dashboard'))
  )
}

function App(){
  const [view,setView]=useState('login')
  const [session,setSession]=useState(null)
  const [plan,setPlan]=useState(null)
  const [showConfetti,setShowConfetti]=useState(false)
  useEffect(()=>{ const username = localStorage.getItem('sc_username'); if(username) api.me().then(u=>{ if(u) setView('dashboard') }) },[])
  function handleLogin(username){ localStorage.setItem('sc_username', username); setView('dashboard') }
  function handleNewAchievements(list){ if(list && list.length>0){ setShowConfetti(true); setTimeout(()=>setShowConfetti(false),4500) } }
  return h('div',{},
    view==='login' ? h(Login,{onLogin:handleLogin}) : null,
    view==='dashboard' ? h(Dashboard,{onStart:()=>setView('create'), onNewAchievements:handleNewAchievements}) : null,
    view==='create' ? h(CreateSession,{onCreated:(s)=>{ setSession(s); setView('player') }}) : null,
    view==='player' && session ? h(SessionPlayer,{session,onDone:(s)=>{
        api.submitResults({session_id:s.id, results:s.tasks.map(t=>({actual_seconds:t.actual_seconds||t.estimated_secs, correct:true}))}).then((pl)=>{ setPlan(pl); setView('plan') }).catch(()=>setView('dashboard'))
    }}) : null,
    view==='plan' && plan ? h(StudyPlan,{plan,onClose:()=>{ setPlan(null); setView('dashboard') }}) : null,
    showConfetti ? h(ConfettiCanvas) : null
  )
}

render(h(App,{}), document.getElementById('root'))
