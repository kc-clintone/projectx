import React, {useState, useEffect, useRef} from 'react'
import clsx from 'clsx'

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

  useEffect(()=>{
    fetch('/ui/mock_tests.json').then(r=>r.json()).then(setMockScenarios)
    fetch('/ui/demo_students.json').then(r=>r.json()).then(ds=>{ const d = ds.find(x=>x.student_id===studentID); if(d) setProfile(d) })
  },[])

  useEffect(()=>{
    if(running) timerRef.current = setInterval(()=> setElapsed(e=>e+1),1000)
    else { clearInterval(timerRef.current); timerRef.current=null }
    return ()=>clearInterval(timerRef.current)
  },[running])

  function createSession(){ const tasks = tasksText.split('\n').filter(Boolean).map((t,i)=>({id:String(i+1),prompt:t})); fetch('/api/v1/session',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({student_id:studentID,student_level:'grade10',subject:'math',tasks})}).then(r=>r.json()).then(setSession) }
  function startSequence(){ if(!session) return; setCurrentIdx(0); setElapsed(0); setRunning(true) }
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
        <div>ID: {profile.student_id}</div>
        <div>Points: {profile.points}</div>
        <div>Streak: {profile.streak}</div>
        <div>Badges: {profile.earned_badges?.map(b=>b.name).join(', ')}</div>
        <div>Recap: {profile.weekly_recap?.map(r=>r.topic+':'+r.improved_by).join(', ')}</div>
        <div>Summary: {profile.summary}</div>
      </div> }
    </div>
  </div>
}
