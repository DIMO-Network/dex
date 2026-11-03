// Logs in with a browser wallet: the wallet signs the challenge the server issues
// for this login, and the server verifies it.
//
// Wallets are found through EIP-6963, with window.ethereum (EIP-1193) as the fallback
// for wallets that don't announce themselves. Nothing here is loaded from another
// origin, so the page's script-src can stay 'self'.

const ERR_USER_REJECTED = 4001;
const ERR_REQUEST_PENDING = -32002;

const announced = new Map();

window.addEventListener("eip6963:announceProvider", (event) => {
  const { info, provider } = event.detail || {};
  if (info && info.uuid && provider && typeof provider.request === "function") {
    announced.set(info.uuid, { info, provider });
  }
});
window.dispatchEvent(new Event("eip6963:requestProvider"));

let challengeUrl;
let verifyUrl;
let authId;
let busy = false;

function wallets() {
  if (announced.size > 0) {
    return [...announced.values()];
  }
  if (window.ethereum && typeof window.ethereum.request === "function") {
    return [{ info: { name: "Browser wallet" }, provider: window.ethereum }];
  }
  return [];
}

function displayError(msg) {
  const errorBox = document.getElementById("web3-error");
  errorBox.style.display = "block";
  errorBox.textContent = msg;
}

function clearError() {
  const errorBox = document.getElementById("web3-error");
  errorBox.style.display = "none";
  errorBox.textContent = "";
}

async function postJson(url, obj) {
  const resp = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(obj),
  });
  let body = {};
  try {
    body = await resp.json();
  } catch (err) {
    // A non-JSON body leaves only the status to report.
  }
  if (!resp.ok) {
    throw new Error(body.message || `Request failed with status ${resp.status}.`);
  }
  return body;
}

// personal_sign takes the message as hex-encoded UTF-8, which every wallet accepts.
function utf8ToHex(str) {
  let hex = "0x";
  for (const b of new TextEncoder().encode(str)) {
    hex += b.toString(16).padStart(2, "0");
  }
  return hex;
}

function walletErrorMessage(err) {
  if (err && err.code === ERR_USER_REJECTED) {
    return "The request was declined in your wallet.";
  }
  if (err && err.code === ERR_REQUEST_PENDING) {
    return "Your wallet already has a request open. Open your wallet to finish it.";
  }
  return (err && err.message) || "Login failed.";
}

async function logIn(provider) {
  if (busy) {
    return;
  }
  busy = true;
  clearError();
  try {
    const accounts = await provider.request({ method: "eth_requestAccounts" });
    if (!accounts || accounts.length === 0) {
      throw new Error("Your wallet did not share an account.");
    }
    const address = accounts[0];

    const challenge = await postJson(challengeUrl, { address, state: authId });
    const signed = await provider.request({
      method: "personal_sign",
      params: [utf8ToHex(challenge.nonce), address],
    });

    const verified = await postJson(verifyUrl, { signed, state: authId });
    window.location.replace(verified.redirect);
  } catch (err) {
    console.error(err);
    displayError(walletErrorMessage(err));
  } finally {
    busy = false;
  }
}

function showWalletChoice(list) {
  const container = document.getElementById("wallet-list");
  container.replaceChildren();
  for (const { info, provider } of list) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "web3-btn web3-wallet-btn";
    // EIP-6963 icons are data URIs. Anything else would load from another origin.
    if (typeof info.icon === "string" && info.icon.startsWith("data:image/")) {
      const icon = document.createElement("img");
      icon.src = info.icon;
      icon.alt = "";
      icon.className = "web3-wallet-icon";
      button.appendChild(icon);
    }
    const label = document.createElement("span");
    label.className = "dex-btn-text web3-btn-text";
    label.textContent = info.name || "Browser wallet";
    button.appendChild(label);
    button.addEventListener("click", () => logIn(provider));
    container.appendChild(button);
  }
  container.style.display = "block";
}

function onConnect() {
  const list = wallets();
  if (list.length === 0) {
    displayError(
      "No browser wallet found. Install a wallet extension, or open this page in your wallet app's browser."
    );
    return;
  }
  if (list.length === 1) {
    logIn(list[0].provider);
    return;
  }
  showWalletChoice(list);
}

function init() {
  challengeUrl = document.getElementById("tmpl-challenge-url").value;
  verifyUrl = document.getElementById("tmpl-verify-url").value;
  authId = document.getElementById("tmpl-auth-id").value;

  document.getElementById("connect-button").addEventListener("click", onConnect);
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", init);
} else {
  init();
}
