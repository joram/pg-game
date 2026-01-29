import { useState, useEffect, useRef } from 'react'
import './App.css'

function App() {
  const [connected, setConnected] = useState(false)
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState('')
  const [chatTurns, setChatTurns] = useState([]) // { id, prompt, response: null | { type, ... } }
  const [history, setHistory] = useState([])
  const [playingId, setPlayingId] = useState(null)
  const nextIdRef = useRef(0)
  const resultsContentRef = useRef(null)
  const wsRef = useRef(null)
  const queryInputRef = useRef(null)
  const retryCountRef = useRef(0)
  const audioRef = useRef(null)

  useEffect(() => {
    // Brief delay so game-server WebSocket has time to start (e.g. when using docker compose up)
    const t = setTimeout(() => {
      connect()
    }, 1500)
    return () => {
      clearTimeout(t)
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [])

  const connect = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.hostname
    const wsUrl = import.meta.env.PROD
      ? `${protocol}//${host}/ws`
      : `${protocol}//${host}:8080/ws`
    
    const ws = new WebSocket(wsUrl)
    
    ws.onopen = () => {
      console.log('Connected to game server')
      retryCountRef.current = 0
      setConnected(true)
      ws.send(JSON.stringify({ type: 'connect' }))
    }
    
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        handleMessage(data)
      } catch (e) {
        setLoading(false)
        appendResponseToLastTurn({ type: 'text', content: event.data })
      }
    }
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      setLoading(false)
      if (retryCountRef.current >= 2) {
        appendResponseToLastTurn({ type: 'error', content: 'Connection error. Is the game server running on port 8080?' })
      }
    }
    
    ws.onclose = () => {
      console.log('Disconnected from game server')
      setConnected(false)
      setLoading(false)
      retryCountRef.current += 1
      setTimeout(() => {
        if (!wsRef.current || wsRef.current.readyState === WebSocket.CLOSED) {
          connect()
        }
      }, 3000)
    }
    
    wsRef.current = ws
  }

  const handleMessage = (data) => {
    setLoading(false)
    if (data.type === 'result') {
      appendResponseToLastTurn({
        type: 'query',
        query: data.query,
        rows: data.rows,
        columns: data.columns,
        rowCount: data.rowCount,
      })
    } else if (data.type === 'error') {
      appendResponseToLastTurn({ type: 'error', content: data.message })
    } else if (data.type === 'text') {
      appendResponseToLastTurn({ type: 'text', content: data.content })
    }
  }

  const appendResponseToLastTurn = (response) => {
    setChatTurns(prev => {
      const next = [...prev]
      const last = next[next.length - 1]
      if (last && last.response === null) {
        next[next.length - 1] = { ...last, response }
      } else {
        next.push({ id: nextIdRef.current++, prompt: null, response })
      }
      return next
    })
    setTimeout(() => resultsContentRef.current?.scrollTo({ top: resultsContentRef.current.scrollHeight, behavior: 'smooth' }), 50)
  }

  const sendQuery = () => {
    if (!query.trim() || !connected || loading) return
    
    let queryText = query.trim()
    if (!queryText.endsWith(';')) {
      queryText += ';'
    }
    
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      setLoading(true)
      setHistory(prev => [queryText, ...prev].filter((q, i, arr) => arr.indexOf(q) === i).slice(0, 50))
      setChatTurns(prev => [...prev, { id: nextIdRef.current++, prompt: queryText, response: null }])
      wsRef.current.send(JSON.stringify({ type: 'query', query: queryText }))
      setQuery('')
      setTimeout(() => resultsContentRef.current?.scrollTo({ top: resultsContentRef.current.scrollHeight, behavior: 'smooth' }), 50)
    }
  }

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendQuery()
    } else if (e.key === 'ArrowUp' && history.length > 0) {
      e.preventDefault()
      const lastQuery = history[0]
      setQuery(lastQuery)
    }
  }

  const speakText = async (text, turnId) => {
    if (!text) return

    // Stop any currently playing audio
    if (audioRef.current) {
      audioRef.current.pause()
      audioRef.current = null
    }

    // If clicking the same one that's playing, just stop
    if (playingId === turnId) {
      setPlayingId(null)
      return
    }

    setPlayingId(turnId)

    try {
      const host = window.location.hostname
      const ttsUrl = import.meta.env.PROD
        ? `${window.location.protocol}//${host}/tts`
        : `${window.location.protocol}//${host}:8080/tts`

      const response = await fetch(ttsUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text })
      })

      if (!response.ok) {
        throw new Error('TTS request failed')
      }

      const blob = await response.blob()
      const url = URL.createObjectURL(blob)
      const audio = new Audio(url)
      audioRef.current = audio

      audio.onended = () => {
        setPlayingId(null)
        URL.revokeObjectURL(url)
        audioRef.current = null
      }

      audio.onerror = () => {
        setPlayingId(null)
        URL.revokeObjectURL(url)
        audioRef.current = null
      }

      await audio.play()
    } catch (error) {
      console.error('TTS error:', error)
      setPlayingId(null)
    }
  }

  const getResponseText = (response) => {
    if (!response) return null
    if (response.type === 'text') return response.content ?? ''
    if (response.type === 'error') return response.content ?? ''
    if (response.type === 'query') {
      if (response.rows?.length > 0 && response.columns?.length > 0) {
        const header = response.columns.join(' | ')
        const rows = response.rows.map(row =>
          row.map(cell => cell !== null ? String(cell) : 'NULL').join(' | ')
        )
        return [header, ...rows].join('\n')
      }
      return `${response.rowCount ?? 0} row(s)`
    }
    return ''
  }

  const formatResponse = (response, turnId) => {
    if (!response) {
      return (
        <div className="chat-response chat-response--pending">
          <span className="spinner" /> Waiting for response...
        </div>
      )
    }

    const fullMessage = getResponseText(response)
    const speakerButton = fullMessage ? (
      <button
        className="speak-button"
        onClick={() => speakText(fullMessage, turnId)}
        title={playingId === turnId ? "Stop" : "Speak"}
        disabled={playingId !== null && playingId !== turnId}
      >
        {playingId === turnId ? '■' : '🔊'}
      </button>
    ) : null

    if (response.type === 'text') {
      return (
        <div className="chat-response chat-response--text">
          <div className="response-content">{response.content}</div>
          {speakerButton}
        </div>
      )
    }
    if (response.type === 'error') {
      return (
        <div className="chat-response chat-response--error">
          <div className="response-content">{response.content}</div>
          {speakerButton}
        </div>
      )
    }
    if (response.type === 'query') {
      if (response.rows && response.rows.length > 0) {
        return (
          <div className="chat-response chat-response--query">
            <table>
              <thead>
                <tr>
                  {response.columns?.map((col, i) => (
                    <th key={i}>{col}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {response.rows.map((row, i) => (
                  <tr key={i}>
                    {row.map((cell, j) => (
                      <td key={j}>{cell !== null ? String(cell) : 'NULL'}</td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
            {speakerButton}
          </div>
        )
      }
      return (
        <div className="chat-response chat-response--query chat-response--empty">
          {response.rowCount ?? 0} row(s)
        </div>
      )
    }
    return null
  }

  return (
    <div className="app">
      <div className="header">
        <h1>Text-Based Adventure Game</h1>
        <div className="connection-status">
          <span className={`status-indicator ${connected ? 'connected' : 'disconnected'}`}>
            {connected ? '●' : '○'}
          </span>
          <span>{connected ? 'Connected' : 'Disconnected'}</span>
        </div>
      </div>
      
      <div className="main-content">
        <div className="query-panel">
          <div className="query-input-container">
            <textarea
              ref={queryInputRef}
              className="query-input"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={handleKeyPress}
              placeholder={connected ? (loading ? "Waiting for response..." : "Enter your command (e.g., 'look', 'take key', 'inventory')...") : "Connecting..."}
              disabled={!connected || loading}
              rows={3}
            />
            <button 
              className="send-button"
              onClick={sendQuery}
              disabled={!connected || loading || !query.trim()}
            >
              {loading ? <span className="send-button-content"><span className="spinner" /> Sending...</span> : 'Send'}
            </button>
          </div>
        </div>
        
        <div className="results-panel">
          <div className="results-header">
            <h2>Chat</h2>
            <button 
              className="clear-button"
              onClick={() => setChatTurns([])}
            >
              Clear
            </button>
          </div>
          <div className="results-content chat-history" ref={resultsContentRef}>
            {chatTurns.length === 0 ? (
              <div className="empty-state">
                <p>No messages yet. Try commands like:</p>
                <ul>
                  <li><code>look</code> - Look around</li>
                  <li><code>inventory</code> - Check your inventory</li>
                  <li><code>take key</code> - Pick up an item</li>
                  <li><code>go north</code> - Move to a new location</li>
                </ul>
              </div>
            ) : (
              chatTurns.map((turn) => (
                <div key={turn.id} className="chat-turn">
                  {turn.prompt != null && (
                    <div className="chat-message chat-message--user">
                      <span className="chat-label">You</span>
                      <div className="chat-bubble chat-bubble--user">{turn.prompt}</div>
                    </div>
                  )}
                  <div className="chat-message chat-message--server">
                    <span className="chat-label">Game</span>
                    <div className={`chat-bubble chat-bubble--server chat-bubble--${turn.response?.type ?? 'pending'} ${playingId === turn.id ? 'chat-bubble--playing' : ''}`}>
                      {playingId === turn.id && (
                        <div className="chat-bubble-playing-overlay">
                          <span className="spinner" />
                        </div>
                      )}
                      {formatResponse(turn.response, turn.id)}
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
