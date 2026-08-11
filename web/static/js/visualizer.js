let simulationData = null;
let currentTurnIndex = 0;
let isPlaying = false;
let animationProgress = 1.0;
let lastTime = 0;

let transform = { x: 0, y: 0, scale: 1 };
let isDragging = false;
let dragStart = { x: 0, y: 0 };

const canvas = document.getElementById('stage');
const ctx = canvas.getContext('2d');

window.addEventListener('resize', resizeCanvas);
resizeCanvas();

// Load list of available maps from Go API
fetchMaps();

function resizeCanvas() {
  canvas.width = canvas.parentElement.clientWidth;
  canvas.height = canvas.parentElement.clientHeight;
  draw();
}

async function fetchMaps() {
  try {
    const res = await fetch('/api/maps');
    const maps = await res.json();
    const select = document.getElementById('map-select');
    select.innerHTML = '';
    
    maps.forEach(map => {
      const opt = document.createElement('option');
      opt.value = map;
      opt.textContent = `Map: ${map}`;
      select.appendChild(opt);
    });

    if (maps.length > 0) {
      loadMapSimulation(maps[0]);
    }
  } catch (err) {
    console.error("Failed to load maps from Go server:", err);
  }
}

async function loadMapSimulation(filename) {
  try {
    const res = await fetch(`/api/simulate?map=${encodeURIComponent(filename)}`);
    if (!res.ok) {
      const errText = await res.text();
      alert(`Go Backend Error: ${errText}`);
      return;
    }
    simulationData = await res.json();
    resetSimulation();
    autoFitGraph();
    play();
  } catch (err) {
    console.error("Error running Go simulation:", err);
  }
}

function resetSimulation() {
  pause();
  currentTurnIndex = 0;
  animationProgress = 1.0;
  updateHUD();
  draw();
}

function autoFitGraph() {
  if (!simulationData || Object.keys(simulationData.rooms).length === 0) return;
  const roomList = Object.values(simulationData.rooms);
  
  let minX = Infinity, maxX = -Infinity;
  let minY = Infinity, maxY = -Infinity;

  roomList.forEach(r => {
    if (r.x < minX) minX = r.x;
    if (r.x > maxX) maxX = r.x;
    if (r.y < minY) minY = r.y;
    if (r.y > maxY) maxY = r.y;
  });

  const width = canvas.width;
  const height = canvas.height;
  const padding = 100;

  const graphW = (maxX - minX) || 1;
  const graphH = (maxY - minY) || 1;

  const scaleX = (width - padding * 2) / graphW;
  const scaleY = (height - padding * 2) / graphH;
  const scale = Math.min(scaleX, scaleY, 70);

  const cx = (minX + maxX) / 2;
  const cy = (minY + maxY) / 2;

  transform.scale = scale;
  transform.x = width / 2 - cx * scale;
  transform.y = height / 2 - cy * scale;
  draw();
}

function draw() {
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  if (!simulationData) return;

  ctx.save();
  ctx.translate(transform.x, transform.y);
  ctx.scale(transform.scale, transform.scale);

  // 1. Draw Tunnels
  ctx.lineWidth = 1.5 / transform.scale;
  ctx.strokeStyle = '#334155';
  simulationData.links.forEach(link => {
    const r1 = simulationData.rooms[link.u];
    const r2 = simulationData.rooms[link.v];
    if (!r1 || !r2) return;

    ctx.beginPath();
    ctx.moveTo(r1.x, r1.y);
    ctx.lineTo(r2.x, r2.y);
    ctx.stroke();
  });

  // 2. Draw Rooms
  const radius = 14 / transform.scale;
  Object.values(simulationData.rooms).forEach(room => {
    ctx.beginPath();
    ctx.arc(room.x, room.y, radius, 0, Math.PI * 2);

    if (room.isStart) {
      ctx.fillStyle = '#059669'; // Emerald
    } else if (room.isEnd) {
      ctx.fillStyle = '#ea580c'; // Coral
    } else {
      ctx.fillStyle = '#334155'; // Slate
    }

    ctx.fill();
    ctx.lineWidth = 2 / transform.scale;
    ctx.strokeStyle = '#475569';
    ctx.stroke();

    ctx.font = `500 ${11 / transform.scale}px Inter`;
    ctx.fillStyle = '#cbd5e1';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';

    let displayName = room.name;
    if (room.isStart) displayName += " (START)";
    if (room.isEnd) displayName += " (END)";

    ctx.fillText(displayName, room.x, room.y - radius * 1.5);
  });

  // 3. Draw Ants
  const antRadius = 7 / transform.scale;
  const prevPos = simulationData.turnStates[Math.max(0, currentTurnIndex - 1)] || {};
  const currPos = simulationData.turnStates[currentTurnIndex] || {};

  const allAntIDs = new Set([...Object.keys(prevPos), ...Object.keys(currPos)]);

  allAntIDs.forEach(idStr => {
    const id = parseInt(idStr);
    const fromRoomName = prevPos[id] || simulationData.startRoom;
    const toRoomName = currPos[id] || simulationData.startRoom;

    const fromRoom = simulationData.rooms[fromRoomName];
    const toRoom = simulationData.rooms[toRoomName];

    if (!fromRoom || !toRoom) return;

    const interpX = fromRoom.x + (toRoom.x - fromRoom.x) * animationProgress;
    const interpY = fromRoom.y + (toRoom.y - fromRoom.y) * animationProgress;

    ctx.beginPath();
    ctx.arc(interpX, interpY, antRadius, 0, Math.PI * 2);
    ctx.fillStyle = '#2563eb';
    ctx.fill();

    ctx.font = `600 ${8 / transform.scale}px 'JetBrains Mono'`;
    ctx.fillStyle = '#ffffff';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(`${id}`, interpX, interpY);
  });

  ctx.restore();
}

function animate(timestamp) {
  if (!lastTime) lastTime = timestamp;
  const delta = timestamp - lastTime;
  lastTime = timestamp;

  const duration = parseInt(document.getElementById('speed-slider').value);

  if (isPlaying && animationProgress < 1.0) {
    animationProgress += delta / duration;
    if (animationProgress >= 1.0) {
      animationProgress = 1.0;
      if (currentTurnIndex < simulationData.turns.length) {
        currentTurnIndex++;
        animationProgress = 0.0;
      } else {
        pause();
      }
      updateHUD();
    }
    draw();
  }

  requestAnimationFrame(animate);
}
requestAnimationFrame(animate);

function play() {
  if (!simulationData || !simulationData.turns || simulationData.turns.length === 0) return;
  if (currentTurnIndex >= simulationData.turns.length) {
    currentTurnIndex = 0;
  }
  isPlaying = true;
  animationProgress = 0.0;
  document.getElementById('btn-play').textContent = '❚❚';
}

function pause() {
  isPlaying = false;
  document.getElementById('btn-play').textContent = '▶';
}

function updateHUD() {
  if (!simulationData) return;
  document.getElementById('stat-total-ants').textContent = simulationData.totalAnts;
  
  const currPos = simulationData.turnStates[currentTurnIndex] || {};
  let completed = 0;
  let active = 0;

  Object.values(currPos).forEach(rName => {
    if (rName === simulationData.endRoom) completed++;
    else active++;
  });

  document.getElementById('stat-active-ants').textContent = active;
  document.getElementById('stat-completed-ants').textContent = completed;
  document.getElementById('stat-turn').textContent = `${currentTurnIndex} / ${simulationData.turns ? simulationData.turns.length : 0}`;

  const progress = (simulationData.turns && simulationData.turns.length > 0) ? (currentTurnIndex / simulationData.turns.length) * 100 : 0;
  document.getElementById('progress-fill').style.width = `${progress}%`;
}

document.getElementById('map-select').addEventListener('change', (e) => {
  loadMapSimulation(e.target.value);
});

document.getElementById('btn-play').addEventListener('click', () => {
  if (isPlaying) pause(); else play();
});

document.getElementById('btn-next').addEventListener('click', () => {
  pause();
  if (simulationData && simulationData.turns && currentTurnIndex < simulationData.turns.length) {
    currentTurnIndex++;
    animationProgress = 1.0;
    updateHUD();
    draw();
  }
});

document.getElementById('btn-prev').addEventListener('click', () => {
  pause();
  if (currentTurnIndex > 0) {
    currentTurnIndex--;
    animationProgress = 1.0;
    updateHUD();
    draw();
  }
});

document.getElementById('progress-bar').addEventListener('click', (e) => {
  if (!simulationData || !simulationData.turns || simulationData.turns.length === 0) return;
  pause();
  const rect = e.currentTarget.getBoundingClientRect();
  const ratio = (e.clientX - rect.left) / rect.width;
  currentTurnIndex = Math.round(ratio * simulationData.turns.length);
  animationProgress = 1.0;
  updateHUD();
  draw();
});

document.getElementById('speed-slider').addEventListener('input', (e) => {
  document.getElementById('speed-val').textContent = `${(e.target.value / 1000).toFixed(1)}s`;
});

document.getElementById('btn-reset-view').addEventListener('click', autoFitGraph);

canvas.addEventListener('mousedown', (e) => {
  isDragging = true;
  dragStart = { x: e.clientX - transform.x, y: e.clientY - transform.y };
});

window.addEventListener('mousemove', (e) => {
  if (isDragging) {
    transform.x = e.clientX - dragStart.x;
    transform.y = e.clientY - dragStart.y;
    draw();
  }
});

window.addEventListener('mouseup', () => { isDragging = false; });

canvas.addEventListener('wheel', (e) => {
  e.preventDefault();
  const zoomFactor = e.deltaY < 0 ? 1.1 : 0.9;
  transform.scale *= zoomFactor;
  draw();
}, { passive: false });
