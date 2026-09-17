// Synthesized reminder chime + system notification. The AudioContext must
// be created inside the start-click gesture or the browser blocks it.

let ctx: AudioContext | null = null

export function primeReminder(wantNotification: boolean): void {
  try {
    ctx ??= new AudioContext()
    if (ctx.state === 'suspended') void ctx.resume()
  } catch {
    ctx = null
  }
  if (wantNotification && typeof Notification !== 'undefined' && Notification.permission === 'default') {
    void Notification.requestPermission()
  }
}

// 880Hz triple chime, synthesized — no audio assets.
export function chime(): void {
  if (!ctx || ctx.state !== 'running') return
  const t0 = ctx.currentTime
  for (let i = 0; i < 3; i++) {
    const start = t0 + i * 0.4
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    osc.frequency.value = 880
    gain.gain.setValueAtTime(0, start)
    gain.gain.linearRampToValueAtTime(0.25, start + 0.02)
    gain.gain.linearRampToValueAtTime(0, start + 0.2)
    osc.connect(gain).connect(ctx.destination)
    osc.start(start)
    osc.stop(start + 0.25)
  }
}

export function notifyOver(activity: string, project: string): void {
  if (typeof Notification === 'undefined' || Notification.permission !== 'granted') return
  try {
    new Notification('已到预期时长', { body: `${activity} / ${project} — 计时继续` })
  } catch {
    // Some platforms throw on construction; the visual state remains.
  }
}
