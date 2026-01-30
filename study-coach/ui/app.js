import { h, render } from 'https://unpkg.com/preact@10.16.0/dist/preact.mjs';
import { useState, useEffect, useRef } from 'https://unpkg.com/preact@10.16.0/hooks/dist/hooks.module.js';

const api = {
  createSession: (body)=>fetch('/api/v1/session',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify(body)}).then(r=>r.json()),
  submitResults: (body)=>fetch('/api/v1/submit_results',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify(body)}).then(r=>r.json()),
  profile: (id)=>fetch(`/api/v1/student/${id}/profile`).then(r=>r.ok?r.json():null)
}

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
    emit(); tick(); setTimeout(()=>{running=false; window.removeEventListener('resize', resize)},4000);
  },[]);
  return h('canvas',{ref:ref,class:'confetti-canvas'});
}

function App(){
  const [studentID,setStudentID] = useState('demo_anna')
  const [tasksText,setTasksText] = useState('5+7\n12*3\nintegrate x^2')
  const [session,setSession] = useState(null)
  const [currentIdx,setCurrentIdx] = useState(0)
  const [running,setRunning] = useState(false)
  const [elapsed,setElapsed] = useState(0)
  const timerRef = useRef(null)
  const [profile,setProfile] = useState(null)
  const [mockScenarios,setMockScenarios] = useState([])

  const [username,setUsername]=useState(localStorage.getItem('sc_username')||'');
  const [password,setPassword]=useState('');
  const [email,setEmail]=useState(localStorage.getItem('sc_email')||'');

  useEffect(()=>{
    fetch('/ui/mock_tests.json').then(r=>r.json()).then(setMockScenarios)
    fetch('/ui/demo_students.json').then(r=>r.json()).then(ds=>{ const d = ds.find(x=>x.student_id===studentID); if(d) setProfile(d) })
  },[])

  useEffect(()=>{
    if(running){ timerRef.current = setInterval(()=> setElapsed(e=>e+1),1000) }
    else { clearInterval(timerRef.current); timerRef.current=null }
    return ()=>clearInterval(timerRef.current)
  },[running])

  function loadMockScenario(i){ const s = mockScenarios[i]; if(!s) return; setTasksText(s.tasks.join('\n')) }
  function loadDemoStudent(i){ fetch('/ui/demo_students.json').then(r=>r.json()).then(ds=>{ setProfile(ds[i]); setStudentID(ds[i].student_id) }) }

  async function createSession(){ const tasks = tasksText.split('\n').filter(Boolean).map((t,i)=>({id:String(i+1),prompt:t})); const sess = await api.createSession({student_id:studentID,student_level:'grade10',subject:'math',tasks}); setSession(sess); setCurrentIdx(0); setElapsed(0); setRunning(false) }

  function startSequence(){ if(!session) return; setCurrentIdx(0); setElapsed(0); setRunning(true) }
  function doneTask(){ if(session && session.tasks[currentIdx]){ session.tasks[currentIdx].actual_seconds = elapsed } setElapsed(0); if(currentIdx+1 >= session.tasks.length){ setRunning(false); promptSubmit() } else { setCurrentIdx(currentIdx+1); setElapsed(0); } }

  async function promptSubmit(){ if(confirm('Submit your answers now?')){ const results = session.tasks.map(t=>({actual_seconds:t.actual_seconds||t.estimated_secs, correct: Math.random()>0.4})); await api.submitResults({session_id:session.id, results}); alert('Study plan generated'); api.profile(studentID).then(setProfile) }}

  function uploadFile(e){ const f = e.target.files[0]; if(!f) return; const reader = new FileReader(); reader.onload = ()=> setTasksText(reader.result.split('\n').slice(0,20).join('\n')); reader.readAsText(f) }
  function uploadImage(e){ alert('image upload demo: image received (not processed in demo)') }

  function login(){ localStorage.setItem('sc_username', username); localStorage.setItem('sc_email', email); if(username) setStudentID(username); api.profile(username).then(p=>{ if(p) setProfile(p); else setProfile({student_id: username, points: 0, streak: 0, earned_badges: [], weekly_recap: []}); }).catch(()=>{ setProfile({student_id: username, points: 0, streak: 0, earned_badges: [], weekly_recap: []}) }) }

  return h('div',{},
    h('div',{style:{display:'flex',gap:12}},
      h('div',{style:{flex:'1 1 320px'}},
         h('div',{class:'card'},
           h('h3',null,'Account / Create Session'),
           h('div',null,
             h('div',null,'Username: ', h('input',{placeholder:'Username',value:username,onInput:e=>setUsername(e.target.value)})),
             h('div',null,'Password: ', h('input',{type:'password',value:password,onInput:e=>setPassword(e.target.value)})),
             h('div',null,'Email (optional): ', h('input',{placeholder:'email@example.com',value:email,onInput:e=>setEmail(e.target.value)})),
             h('div',null, h('button',{onClick:login},'Login / Save locally'))
           ),

           h('hr',null),

           h('div',null,'Student ID: ', h('input',{value:studentID,onInput:e=>setStudentID(e.target.value)})),
           h('div',null,'Tasks:'),
           h('textarea',{style:{width:'100%',height:120},value:tasksText,onInput:e=>setTasksText(e.target.value)}),
           h('div',{class:'uploader'},
             h('label',{class:'card',style:{padding:8,cursor:'pointer'}}, h('div',null,'Upload File'), h('input',{type:'file',onChange:uploadFile,style:{display:'none'}})),
             h('label',{class:'card',style:{padding:8,cursor:'pointer'}}, h('div',null,'Upload Image'), h('input',{type:'file',accept:'image/*',onChange:uploadImage,style:{display:'none'}})),
             h('button',{onClick:()=>{ navigator.clipboard.readText().then(t=>setTasksText(t)).catch(()=>alert('Paste not available')) }},'Paste Text')
           ),
           h('div',{style:{marginTop:8}}, h('button',{onClick:createSession},'Create Session'), ' ', h('button',{onClick:()=>loadMockScenario(0)},'Load Quick Math'), ' ', h('button',{onClick:()=>loadMockScenario(1)},'Load Reading'))
         )
      ),
      h('div',{style:{flex:2}},
        h('div',{class:'card'},
          h('h3',null,'Session Player'),
          session ? h('div',null, h('div',null,'Session: ', session.id), h('div',null, 'Current task: ', (session.tasks[currentIdx]||{}).prompt || 'none'), h('div',{class:'taskLarge'}, elapsed+'s elapsed'), h('div',null, h('button',{onClick:doneTask},'Done'), ' ', h('button',{onClick:startSequence},'Start Sequential')) ) : h('div',null,'No session yet')
        ),
        h('div',{style:{height:12}}),
        h('div',{class:'card'}, h('h3',null,'Demo Students'), h('div',null, h('button',{onClick:()=>loadDemoStudent(0)},'Load Anna'), ' ', h('button',{onClick:()=>loadDemoStudent(1)},'Load Ben')))
      )
    ),
    h('div',{style:{marginTop:12}}, profile ? h('div',{class:'card'}, h('h3',null,'Profile'), h('div',null,'ID: ', profile.student_id), h('div',null,'Points: ', profile.points), h('div',null,'Streak: ', profile.streak), h('div',null, 'Badges: ', profile.earned_badges.map(b=>b.name).join(', ')), h('div',null,'Recap: ', profile.weekly_recap.map(r=>r.topic+':'+r.improved_by).join(', ')) ) : null ),
    h(ConfettiCanvas)
  )
}

render(h(App,{}), document.getElementById('root'))
