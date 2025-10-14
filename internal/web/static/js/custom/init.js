const BOX_SIZE = 248;
const DOT_SIZE = 20;
const MAX_VALUE = 32767; // joystick max value
let axisX = 0;
let axisY = 0;

// Ensure Alpine is loaded
document.addEventListener('alpine:init', () => {
    // Global store to hold profiles and connection state
    Alpine.store('profiles', {
        activeProfile: null,
        selectedProfile: null,
    });

    // Create a single SSE connection
    const sse = new EventSource('/events');

    sse.onopen = () => {
        console.log('SSE connected');
    };

    sse.onerror = () => {
        console.log('SSE disconnected, retrying...');
    };

    sse.onmessage = e => console.log('default message', e.data);

    // Listen for custom events from server
    sse.addEventListener('activeProfile', ev => {
        try {
            Alpine.store('profiles').activeProfile = ev.data;
        } catch (e) {
            console.error('Invalid activeProfile event', e);
        }
    });

    sse.addEventListener('selectedProfile', ev => {
        try {
            Alpine.store('profiles').selectedProfile = ev.data;
        } catch (e) {
            console.error('Invalid selectedProfile event', e);
        }
    });

    sse.addEventListener('joystick', ev => {
        const data = JSON.parse(ev.data);

        switch (data.type) {
            case "button":
                handleButton(data)
                break;
            case "hat":
                handleHat(data)
                break;
            case "axis":
                handleAxis(data)
                break;
        }
    })

    // Optional cleanup when page unloads
    window.addEventListener('beforeunload', () => sse.close());
});

function updateDeadzone(val) {
    let value = document.getElementById("deadzoneSlider").value
    deadzone = value
    document.getElementById("deadzoneText").innerText = value
}


function handleButton(data) {
    const key = `[data-key="button-${data.index}"]`;

    const els = document.querySelectorAll(key);
    if (!els || els.length == 0) return;

    els.forEach(function (el) {
        if (data.value > 0) el.classList.add("pressed");
        else el.classList.remove("pressed");
    })
}


function handleHat(data) {
    var key
    switch (data.value) {
        case 1:
            key = `[data-key="hat-${data.index}-up"]`;
            break;
        case 2:
            key = `[data-key="hat-${data.index}-right"]`;
            break;
        case 4:
            key = `[data-key="hat-${data.index}-down"]`;
            break;
        case 8:
            key = `[data-key="hat-${data.index}-left"]`;
            break;
        default:
            key = `[data-key^="hat-${data.index}-"]`;
            break;
    }

    const els = document.querySelectorAll(key);
    if (!els || els.length == 0) return;

    els.forEach(function (el) {
        if (data.value > 0) el.classList.add("pressed");
        else el.classList.remove("pressed");
    })
}

function handleAxis(data) {
    switch (data.index) {
        case 0:
            axisX = data.value || 0;
            break;
        case 1:
            axisY = data.value || 0;
            break;
    }

    const key = `[data-key^="axis-${data.index}-"]`;

    const els = document.querySelectorAll(key);
    if (!els || els.length == 0) return;

    els.forEach(function (el) {
        if (
            axisX >= deadzone &&
            el.getAttribute("data-key").endsWith("0-positive")
        ) {
            el.classList.add("pressed");
        } else if (
            axisY >= deadzone &&
            el.getAttribute("data-key").endsWith("1-positive")
        ) {
            el.classList.add("pressed");
        } else if (
            axisX <= -deadzone &&
            el.getAttribute("data-key").endsWith("0-negative")
        ) {
            el.classList.add("pressed");
        } else if (
            axisY <= -deadzone &&
            el.getAttribute("data-key").endsWith("1-negative")
        ) {
            el.classList.add("pressed");
        } else {
            el.classList.remove("pressed");
        }
    })

    const dot = document.getElementById("joystick-dot")

    const maxRadius = 124;

    // Map int16 (-32768 -> 32767) to -maxRadius -> +maxRadius
    const x = Math.max(Math.min(axisX / 32767 * maxRadius, maxRadius), -maxRadius);
    const y = Math.max(Math.min(axisY / 32767 * maxRadius, maxRadius), -maxRadius);

    dot.style.left = `calc(50% + ${x}px)`;
    dot.style.top = `calc(50% + ${y}px)`;

    if (Math.abs(axisX) < 500 && Math.abs(axisY) < 500) {
        dot.setAttribute("hidden", "hidden")
    } else {
        dot.removeAttribute("hidden")
    }
}


