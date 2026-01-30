import React, {useState, useEffect, useRef} from 'react'
import clsx from 'clsx'

/**
 * Badge renders an inline SVG badge and applies a subtle animation when new.
 * @param {{name:string, icon?:string}} props
 */
function Badge({name, icon}){
  return <div className="badge" title={name} aria-hidden>
    <svg viewBox="0 0 64 64" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <linearGradient id={`g-${name}`} x1="0" x2="1">
          <stop offset="0%" stopColor="#ffd54a"/>
          <stop offset="100%" stopColor="#ffb74d"/>
        </linearGradient>
      </defs>
      <circle cx="32" cy="32" r="30" fill="#fff3e0"/>
      <path d="M32 10 L39 28 L58 28 L42 38 L48 56 L32 46 L16 56 L22 38 L6 28 L25 28 Z" fill={`url(#g-${name})`} stroke="#ff9800" strokeWidth="1" className="badge-path" />
    </svg>
  </div>
}

/** Confetti component uses a canvas to render a short burst of confetti particles. */
function Confetti({trigger}){
  const canvasRef = useRef(null)
  useEffect(()=>{
    if(!trigger) return;
    const canvas = canvasRef.current; if(!canvas) return;
    const ctx = canvas.getContext('2d');
    let running = true;
    function resize(){ canvas.width = window.innerWidth; canvas.height = window.innerHeight }
    resize(); window.addEventListener('resize', resize);
    const particles = [];
    function emit(){ for(let i=0;i<80;i++){ particles.push({x:Math.random()*canvas.width,y:-10, vx:(Math.random()-0.5)*8, vy:Math.random()*6+2, life:Math.random()*80+60, color:['#f59e0b','#ef4444','#10b981','#3b82f6'][Math.floor(Math.random()*4)], size:Math.random()*8+4}) } }
    function tick(){ ctx.clearRect(0,0,canvas.width,canvas.height); for(let i=particles.length-1;i>=0;i--){ const p=particles[i]; p.x+=p.vx; p.y+=p.vy; p.vy+=0.25; p.life--; ctx.fillStyle=p.color; ctx.fillRect(p.x,p.y,p.size,p.size); if(p.life<=0||p.y>canvas.height+20) particles.splice(i,1) } if(running) requestAnimationFrame(tick) }
    emit(); tick(); setTimeout(()=>{ running=false; window.removeEventListener('resize', resize) }, 3500);
  },[trigger])
  return <canvas ref={canvasRef} className="confetti-canvas" aria-hidden="true" />
}

// App is the main frontend component
export default function App(){
  const [studentID, setStudentID] = useState('demo_anna')
  const [tasksText, setTasksText] = useState('5+7\n12*3\nintegrate x^2')
  const [session, setSession] = useState(null)
  const [currentIdx, setCurrentIdx] = useState(0)
  const [running, setRunning] = useState(false)
  const [elapsed, setElapsed] = useState(0)
  const timerRef = useRef(null)
  const [profile, setProfile] = useState(null)
  const [mockScenarios, setMockScenarios] = useState([])
  const [confettiTrigger, setConfettiTrigger] = useState(false)
  const prevBadgesRef = useRef(0)

  useEffect(()=>{
    fetch('/ui/mock_tests.json').then(r=>r.json()).then(setMockScenarios)
    fetch('/ui/demo_students.json').then(r=>r.json()).then(ds=>{ const d = ds.find(x=>x.student_id===studentID); if(d) setProfile(d) })
  },[])

  useEffect(()=>{
    if(running) timerRef.current = setInterval(()=> setElapsed(e=>e+1),1000)
    else { clearInterval(timerRef.current); timerRef.current=null }
    return ()=>clearInterval(timerRef.current)
  },[running])

  // detect new badges and trigger confetti
  useEffect(()=>{
    const count = profile?.earned_badges?.length || 0
    if(prevBadgesRef.current && count > prevBadgesRef.current){
      setConfettiTrigger(c=>!c)
    }
    prevBadgesRef.current = count
  },[profile?.earned_badges])

  /** createSession posts the tasks to the backend and stores the generated session */
  function createSession(){ const tasks = tasksText.split('\n').filter(Boolean).map((t,i)=>({id:String(i+1),prompt:t})); fetch('/api/v1/session',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({student_id:studentID,student_level:'grade10',subject:'math',tasks})}).then(r=>r.json()).then(setSession) }
  /** startSequence begins sequential timed practice */
  function startSequence(){ if(!session) return; setCurrentIdx(0); setElapsed(0); setRunning(true) }
  /** doneTask marks current task done and steps to the next */
  function doneTask(){ if(session && session.tasks[currentIdx]) session.tasks[currentIdx].actual_seconds = elapsed; setElapsed(0); if(currentIdx+1 >= session.tasks.length){ setRunning(false); promptSubmit() } else { setCurrentIdx(currentIdx+1); setElapsed(0) } }
  async function promptSubmit(){ if(confirm('Submit your answers now?')){ const results = session.tasks.map(t=>({actual_seconds:t.actual_seconds||t.estimated_secs, correct: Math.random()>0.4})); await fetch('/api/v1/submit_results',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({session_id:session.id, results})}); alert('Study plan generated'); fetch(`/api/v1/student/${studentID}/profile`).then(r=>r.json()).then(setProfile) }}

  return <div className="app">
    <div className="row">
      <div style={{flex:'1 1 320px'}}>
        <div className="card">
          <h3>Account / Create Session</h3>
          <div>Student ID: <input value={studentID} onChange={e=>setStudentID(e.target.value)} /></div>
          <div>Tasks:</div>
          <textarea style={{width:'100%',height:120}} value={tasksText} onChange={e=>setTasksText(e.target.value)} />
          <div className="uploader">
            <label className="card" style={{padding:8,cursor:'pointer'}}>Upload File <input type="file" onChange={e=>{ const f=e.target.files[0]; const reader=new FileReader(); reader.onload=()=>setTasksText(reader.result.slice(0,2000)); reader.readAsText(f); }} style={{display:'none'}} /></label>
            <label className="card" style={{padding:8,cursor:'pointer'}}>Upload Image <input type="file" accept="image/*" onChange={()=>alert('image demo')} style={{display:'none'}} /></label>
            <button onClick={()=>navigator.clipboard.readText().then(t=>setTasksText(t)).catch(()=>alert('Paste not available'))}>Paste Text</button>
          </div>
          <div style={{marginTop:8}}><button onClick={createSession}>Create Session</button> <button onClick={()=>setTasksText(mockScenarios[0]?.tasks.join('\n')||'')}>Load Quick Math</button></div>
        </div>
      </div>
      <div style={{flex:2}}>
        <div className="card">
          <h3>Session Player</h3>
          { session ? <div>
            <div>Session: {session.id}</div>
            <div>Current task: {session.tasks[currentIdx]?.prompt || 'none'}</div>
            <div className="taskLarge">{elapsed}s elapsed</div>
            <div><button onClick={doneTask}>Done</button> <button onClick={startSequence}>Start Sequential</button></div>
          </div> : <div>No session yet</div> }
        </div>
        <div style={{height:12}} />
        <div className="card">
          <h3>Demo Students</h3>
          <div><button onClick={()=>fetch('/ui/demo_students.json').then(r=>r.json()).then(ds=>{ setProfile(ds[0]); setStudentID(ds[0].student_id) })}>Load Anna</button></div>
        </div>
      </div>
    </div>
    <div style={{marginTop:12}}>
      { profile && <div className="card">
        <h3>Profile</h3>
        <div style={{display:'flex',alignItems:'center'}}>
          <div style={{flex:1}}>
            <div>ID: {profile.student_id}</div>
            <div>Points: {profile.points}</div>
            <div>Streak: {profile.streak}</div>
            <div>Recap: {profile.weekly_recap?.map(r=>r.topic+':'+r.improved_by).join(', ')}</div>
            <div style={{marginTop:8}}>Summary: {profile.summary}</div>
          </div>
          <div style={{marginLeft:12}}>
            {profile.earned_badges?.map(b=> <div key={b.id} style={{display:'flex',alignItems:'center'}}><Badge name={b.name} /> <div style={{marginLeft:8}}>{b.name}</div></div>)}
          </div>
        </div>
      </div> }
    </div>
    <Confetti trigger={confettiTrigger} />
  </div>
}
