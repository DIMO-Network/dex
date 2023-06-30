import {
  EthereumClient,
  w3mConnectors,
  w3mProvider,
  WagmiCore,
  WagmiCoreChains,
  WagmiCoreConnectors,
} from "https://unpkg.com/@web3modal/ethereum@2.6.2";

import { Web3Modal } from "https://unpkg.com/@web3modal/html@2.6.2";

// 0. Import wagmi dependencies
const { mainnet, polygon, polygonMumbai, avalanche, arbitrum } =
  WagmiCoreChains;
const {
  configureChains,
  createConfig,
  switchNetwork,
  disconnect,
  signMessage,
  getAccount,
} = WagmiCore;

// 1. Define chains
const chains = [mainnet, polygon, polygonMumbai];
const projectId = "3d6930929763b3142513d912505eba46";

// 2. Configure wagmi client
const { publicClient } = configureChains(chains, [w3mProvider({ projectId })]);

const wagmiConfig = createConfig({
  autoConnect: true,
  connectors: [
    ...w3mConnectors({ chains, version: 2, projectId }),
    new WagmiCoreConnectors.CoinbaseWalletConnector({
      chains,
      options: {
        appName: "html wagmi example",
      },
    }),
  ],
  publicClient,
});

// 3. Create ethereum and modal clients
const ethereumClient = new EthereumClient(wagmiConfig, chains);
export const web3Modal = new Web3Modal(
  {
    projectId,
    walletImages: {
      safe: "https://pbs.twimg.com/profile_images/1566773491764023297/IvmCdGnM_400x400.jpg",
    },
    enableAccountView: true,
    enableExplorer: true,
    defaultChain: polygon,
    themeMode: "dark",
    themeVariables: {
      '--w3m-accent-color': '#FF8700',
      '--w3m-accent-fill-color': '#000000',
      '--w3m-background-color': '#000000',
      // '--w3m-background-image-url': '/images/customisation/background.png',
      // '--w3m-logo-image-url': '/images/customisation/logo.png',
      '--w3m-background-border-radius': '0px',
      '--w3m-container-border-radius': '0px',
      '--w3m-wallet-icon-border-radius': '0px',
      '--w3m-wallet-icon-large-border-radius': '0px',
      '--w3m-input-border-radius': '0px',
      '--w3m-button-border-radius': '0px',
      '--w3m-secondary-button-border-radius': '0px',
      '--w3m-notification-border-radius': '0px',
      '--w3m-icon-button-border-radius': '0px',
      '--w3m-button-hover-highlight-border-radius': '0px',
      '--w3m-font-family': 'monospace'
    }
  },
  ethereumClient
);

async function postJson(url, obj) {
  return fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(obj),
  });
}

// web3Modal.subscribeModal(async (newState) => {
//   const { open } = newState;
//   const account = await getAccount();
//   if (open && account.address) {
//     onAccountConnected();
//   }
// });

function displayError(msg) {
  const errorBox = document.getElementById("web3-error");
  errorBox.style.display = "block";
  errorBox.textContent = msg;
}

async function onAccountConnected(newState) {
  try {
    const account = await getAccount();
    console.log(account.address);
    const challengeResp = await postJson(challengeUrl, {
      address: account.address,
      state: authId,
    });
    const challengeBody = await challengeResp.json();
    //console.log(challengeBody);
    const signature = await signMessage({
      message: challengeBody.nonce,
    });
    //Catch denied error here and fail gracefully.
    console.log(signature);
    const verifyResp = await postJson(verifyUrl, { signed: signature, state: authId });
    if (verifyResp.ok) {
      const verifyBody = await verifyResp.json();
      window.location.replace(verifyBody.redirect);
    } else {
      const verifyBody = await verifyResp.json();
      displayError("Verification error: " + verifyBody.message);
    }
  } catch (err) {
    console.error(err);
    // await onReset();
  }
}

//4. Create empty variables
let infuraId;
let challengeUrl;
let authId;
let verifyUrl;

//5. Init Function
async function init() {
  infuraId = document.getElementById("tmpl-infura-id").value;
  challengeUrl = document.getElementById("tmpl-challenge-url").value;
  authId = document.getElementById("tmpl-auth-id").value;
  verifyUrl = document.getElementById("tmpl-verify-url").value;

  const connectButton = document.getElementById("connect-button");
  connectButton.addEventListener("click", onConnect);

  const resetButton = document.getElementById("reset-button");
  resetButton.addEventListener("click", onReset);

  await web3Modal.subscribeEvents(async (newState) => {
    console.log("EV " + JSON.stringify(newState));
    const { name } = newState;
    if (name === "ACCOUNT_CONNECTED") {
      await onAccountConnected(newState);
    }
    if (name === "ACCOUNT_DISCONNECTED") {
      await onReset();
    }
  });
  // web3Modal.openModal();
}

// 4. Connect
async function onConnect() {
  try {
    await web3Modal.openModal();
  } catch (err) {
    console.error(err);
    await onReset();
  }
}

async function onReset() {
  try {
    await disconnect()
    await web3Modal.closeModal();
  } catch (err) {
    console.error(err);
  }
}

window.addEventListener("load", init);
