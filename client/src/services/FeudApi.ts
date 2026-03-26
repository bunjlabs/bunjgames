import GameApi from './GameApi';

const FEUD_TOKEN = 'FEUD_TOKEN';
const FEUD_PLAYER_NAME = 'FEUD_PLAYER_NAME';

export default class FeudApi extends GameApi {
  playerName: string | null = null;

  constructor(apiEndpoint: string, wsEndpoint: string) {
    super(apiEndpoint, wsEndpoint, FEUD_TOKEN);
    this.loadPlayerName();
  }

  createGame(inputFile: HTMLInputElement) {
    const formData = new FormData();
    formData.append('game', 'feud');
    formData.append('file', inputFile.files![0]);

    return this.axios
      .post('create', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      .then((result) => {
        this.saveToken(result.data.token);
        return result.data;
      });
  }

  registerPlayer(token: string, name: string) {
    return this.axios
      .post('register', { token, name })
      .then((result) => {
        this.saveToken(result.data.game.token);
        this.savePlayerName(result.data.player);
        return result.data;
      });
  }

  calcTime(): number {
    const now = Date.now();
    const time = Math.round((this.game!.timer - now) / 1000);
    return Math.clamp(time, Number.MAX_VALUE, 0);
  }

  loadPlayerName(): string | null {
    try {
      this.playerName = JSON.parse(
        localStorage.getItem(FEUD_PLAYER_NAME) ?? 'null',
      );
    } catch {
      this.playerName = null;
    }
    return this.playerName;
  }

  savePlayerName(name: string) {
    this.playerName = name;
    localStorage.setItem(FEUD_PLAYER_NAME, JSON.stringify(this.playerName));
  }

  nextState(fromState: string | null = null) {
    this.execute('next', { from: fromState });
  }

  buttonClick(playerName: string) {
    this.execute('buttonClick', { player: playerName });
  }

  setAnswerer(playerName: string) {
    this.execute('setAnswerer', { player: playerName });
  }

  answer(isCorrect: boolean, answerIndex: number) {
    this.execute('answer', {
      correct: Boolean(isCorrect),
      answerIndex,
    });
  }

  hasPlayerName(): boolean {
    return Boolean(this.playerName);
  }
}
