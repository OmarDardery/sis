
const SECRET_KEY = window.FRONTEND_SECRET; // same as env FRONTEND_SECRET_KEY

// ====================== SIGN IN ======================
async function signIn(role) {
  const email = document.getElementById("email").value;
  const password = document.getElementById("password").value;

  const res = await fetch(`/signin/${role}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password, secret_key: SECRET_KEY })
  });

  const data = await res.json();
  if (res.ok) {
    localStorage.setItem("jwt", data.token);
    localStorage.setItem("role", role);
    if (role === "student") window.location.href = "/student/home";
    else window.location.href = "/teacher/home";
  } else {
    alert(data.error || "Sign in failed");
  }
}

// ====================== SIGN UP ======================
async function signUp(role) {
  const email = document.getElementById("email").value;
  const password = document.getElementById("password").value;

  const res = await fetch(`/signup/${role}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password, secret_key: SECRET_KEY })
  });

  const data = await res.json();
  if (res.ok) {
    alert("Signup successful! You can now log in.");
    window.location.href = "/signin";
  } else {
    alert(data.error || "Signup failed");
  }
}

// ====================== TEACHER QR CREATION ======================
async function openAttendance(slotId) {
  const token = localStorage.getItem("jwt");
  const socket = new WebSocket(`wss://${window.location.host}/ws/attendance`);

  socket.onopen = () => {
    socket.send(JSON.stringify({ teacher_token: token, slot_id: slotId })); // teacher_id extracted from JWT ideally
  };

  socket.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    if (msg.slot_id) {
      generateQRCode(msg.slot_id);
    }
  };
}

function generateQRCode(slotId) {
  const qrContainer = document.getElementById("qr-container");
  qrContainer.innerHTML = "";
  const qr = new QRCode(qrContainer, {
    text: JSON.stringify({ slot_id: slotId }),
    width: 200,
    height: 200
  });
}

// ====================== STUDENT QR READER ======================
async function startQRReader() {
  const video = document.createElement("video");
  document.getElementById("reader").appendChild(video);
  const stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: "environment" } });
  video.srcObject = stream;
  video.setAttribute("playsinline", true);
  video.play();

  const canvas = document.createElement("canvas");
  const ctx = canvas.getContext("2d");

  setInterval(async () => {
    if (video.readyState === video.HAVE_ENOUGH_DATA) {
      canvas.height = video.videoHeight;
      canvas.width = video.videoWidth;
      ctx.drawImage(video, 0, 0, canvas.width, canvas.height);
      const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
      const code = jsQR(imageData.data, canvas.width, canvas.height);

      if (code) {
        const data = JSON.parse(code.data);
        await markAttendance(data.slot_id);
        stream.getTracks().forEach(track => track.stop());
      }
    }
  }, 1000);
}

async function markAttendance(slotId) {
  const token = localStorage.getItem("jwt");
  const res = await fetch(`/attendance/mark`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${token}`
    },
    body: JSON.stringify({
      slot_id: slotId,
      secret_key: SECRET_KEY
    })
  });
  const data = await res.json();
  alert(data.message || data.error);
}