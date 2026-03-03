import axios from 'axios';
import Subscriber from "subscriber";
import { toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

export default class GameApi {
    constructor(apiEndpoint, wsEndpoint, tokenName) {
        this.axios = axios.create({
            baseURL: apiEndpoint,
            timeout: 300000
        });

        this.wsEndpoint = wsEndpoint;
        if (!wsEndpoint.startsWith("ws://") && !wsEndpoint.startsWith("wss://")) {
            this.wsEndpoint = ((window.location.protocol === "https:") ? "wss://" : "ws://") + window.location.host + wsEndpoint;
        }
        this.gameSubscriber = new Subscriber();
        this.stateSubscriber = new Subscriber();
        this.intercomSubscriber = new Subscriber();
        this.lastState = null;
        this.tokenName = tokenName;

        this.loadToken();
    }

    connect(token = this.token, checkGame) {
        return new Promise((resolve, reject) => {
            let timeout = setTimeout(() => reject(), 5000);
            let reconnectCount = 5;
            let connected = false;

            const doConnect = () => {
                this.socket = new WebSocket(this.wsEndpoint + token);
                console.log("[WS] Connecting as", token);
                this.socket.onopen = e => {
                    console.log("[WS] Connected", e);
                    this.saveToken(token);
                }
                this.socket.onmessage = e => {
                    const data = JSON.parse(e.data);
                    console.log("[WS] Message", data);
                    if (timeout) {
                        clearTimeout(timeout);
                        if(checkGame && !checkGame(data.message)) {
                            this.socket.close();
                            reject();
                        } else {
                            connected = true;
                            reconnectCount = 5;
                            resolve();
                        }
                    }
                    this.onData(data);
                }
                this.socket.onclose = e => {
                    console.log("[WS] Close", e);

                    if(!connected || !this.hasToken()) {
                        reject();
                    } else {
                        console.log("[WS] Reconnecting", e);

                        if(reconnectCount > 0) {
                            reconnectCount -= 1;
                            setTimeout(() => doConnect(), 1000);
                        }
                    }
                }
                this.socket.onerror = e => {
                    console.log("[WS] Error", e);
                    reject();
                }
            }

            doConnect();
        });
    }

    isConnected() {
        return Boolean(this.socket && this.socket.readyState === WebSocket.OPEN);
    }

    onData(data) {
        if (!data || !data.type) return;

        if (data.type === "game") {
            this.game = data.message;
            if (this.lastState !== this.game.state) {
                this.lastState = this.game.state;
                this.stateSubscriber.fire(this.game);
            }
            this.gameSubscriber.fire(this.game);
        } else if (data.type === "intercom") {
            this.intercomSubscriber.fire(data.message);
        } else if (data.type === "error") {
            toast.dark(data.message);
        }
    }

    loadToken() {
        try {
            this.token = JSON.parse(localStorage.getItem(this.tokenName));
        } catch (e) {
            this.token = null;
        }
        return this.token;
    }

    saveToken(token) {
        this.token = token;
        localStorage.setItem(this.tokenName, JSON.stringify(this.token));
    }

    getGameSubscriber() {
        return this.gameSubscriber;
    }

    getStateSubscriber() {
        return this.stateSubscriber;
    }


    getIntercomSubscriber() {
        return this.intercomSubscriber;
    }

    execute(method, params = {}) {
        if (!this.isConnected()) return;

        console.log("Execute", method, params);

        this.socket.send(JSON.stringify({
            method: method,
            params: params
        }));
    }

    intercom(message) {
        if (!this.isConnected()) return;

        console.log("Intercom", message);

        this.socket.send(JSON.stringify({
            method: "intercom",
            message: message
        }));
    }

    hasToken() {
        return Boolean(this.token);
    }

    setSavedUsername(username) {
        localStorage.setItem("username", JSON.stringify(username));
    }

    getSavedUsername() {
        try {
            return JSON.parse(localStorage.getItem("username"));
        } catch (e) {
            return null;
        }
    }

    logout() {
        this.saveToken(null);
        this.game = undefined;
        if (this.isConnected()) this.socket.close();
    }
}
