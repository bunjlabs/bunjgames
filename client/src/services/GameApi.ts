import axios, { AxiosInstance } from 'axios';
import Subscriber from './Subscriber';
import { toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

export interface GameState {
  token: string;
  state: any;
  players?: any[];
  [key: string]: any;
}

export default class GameApi {
  protected axios: AxiosInstance;
  protected wsEndpoint: string;
  protected gameSubscriber = new Subscriber<GameState>();
  protected stateSubscriber = new Subscriber<GameState>();
  protected intercomSubscriber = new Subscriber<string>(false);
  protected socket: WebSocket | null = null;
  protected lastState: string | null = null;
  protected tokenName: string;

  token: string | null = null;
  game: GameState | undefined;

  constructor(apiEndpoint: string, wsEndpoint: string, tokenName: string) {
    this.axios = axios.create({
      baseURL: apiEndpoint,
      timeout: 300000,
    });

    this.wsEndpoint = wsEndpoint;
    if (!wsEndpoint.startsWith('ws://') && !wsEndpoint.startsWith('wss://')) {
      this.wsEndpoint =
        (window.location.protocol === 'https:' ? 'wss://' : 'ws://') +
        window.location.host +
        wsEndpoint;
    }

    this.tokenName = tokenName;
    this.loadToken();
  }

  connect(
    token: string | null = this.token,
    checkGame?: (game: GameState) => boolean,
  ): Promise<void> {
    return new Promise((resolve, reject) => {
      let timeout: ReturnType<typeof setTimeout> | null = setTimeout(
        () => reject(),
        5000,
      );
      let reconnectCount = 5;
      let connected = false;

      const doConnect = () => {
        this.socket = new WebSocket(this.wsEndpoint + token);

        this.socket.onopen = () => {
          this.saveToken(token);
        };

        this.socket.onmessage = (e) => {
          const data = JSON.parse(e.data);
          if (timeout) {
            clearTimeout(timeout);
            timeout = null;
            if (checkGame && !checkGame(data.message)) {
              this.socket?.close();
              reject();
            } else {
              connected = true;
              reconnectCount = 5;
              resolve();
            }
          }
          this.onData(data);
        };

        this.socket.onclose = () => {
          if (!connected || !this.hasToken()) {
            reject();
          } else if (reconnectCount > 0) {
            reconnectCount -= 1;
            setTimeout(() => doConnect(), 1000);
          }
        };

        this.socket.onerror = () => {
          reject();
        };
      };

      doConnect();
    });
  }

  isConnected(): boolean {
    return Boolean(this.socket && this.socket.readyState === WebSocket.OPEN);
  }

  private onData(data: { type: string; message: any }) {
    if (!data?.type) return;

    if (data.type === 'game') {
      this.game = data.message;
      if (this.lastState !== this.game!.state?.value && this.lastState !== this.game!.state) {
        this.lastState = this.game!.state?.value ?? this.game!.state;
        this.stateSubscriber.fire(this.game!);
      }
      this.gameSubscriber.fire(this.game!);
    } else if (data.type === 'intercom') {
      this.intercomSubscriber.fire(data.message);
    } else if (data.type === 'error') {
      toast.dark(data.message);
    }
  }

  loadToken(): string | null {
    try {
      this.token = JSON.parse(localStorage.getItem(this.tokenName) ?? 'null');
    } catch {
      this.token = null;
    }
    return this.token;
  }

  saveToken(token: string | null) {
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

  execute(method: string, params: Record<string, any> = {}) {
    if (!this.isConnected()) return;
    this.socket!.send(JSON.stringify({ type: method, message: params }));
  }

  intercom(message: string) {
    if (!this.isConnected()) return;
    this.socket!.send(JSON.stringify({ type: 'intercom', message }));
  }

  hasToken(): boolean {
    return Boolean(this.token);
  }

  setSavedUsername(username: string) {
    localStorage.setItem('username', JSON.stringify(username));
  }

  getSavedUsername(): string | null {
    try {
      return JSON.parse(localStorage.getItem('username') ?? 'null');
    } catch {
      return null;
    }
  }

  logout() {
    this.saveToken(null);
    this.game = undefined;
    if (this.isConnected()) this.socket!.close();
  }
}
