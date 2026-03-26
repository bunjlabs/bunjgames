import GameApi from './GameApi';

const WEAKEST_TOKEN = 'WEAKEST_TOKEN';
const WEAKEST_PLAYER_NAME = 'WEAKEST_PLAYER_NAME';

export default class WeakestApi extends GameApi {
  playerName: string | null = null;

  constructor(apiEndpoint: string, wsEndpoint: string) {
    super(apiEndpoint, wsEndpoint, WEAKEST_TOKEN);
    this.loadPlayerName();
  }

  createGame(inputFile: HTMLInputElement) {
    const formData = new FormData();
    formData.append('game', 'weakest');
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

  getTime(from: number): number {
    const serverTime = this.game!.roundState.time;
    const now = Date.now();
    return Math.max(Math.floor((serverTime - (now - from)) / 1000), 0);
  }

  loadPlayerName(): string | null {
    try {
      this.playerName = JSON.parse(
        localStorage.getItem(WEAKEST_PLAYER_NAME) ?? 'null',
      );
    } catch {
      this.playerName = null;
    }
    return this.playerName;
  }

  savePlayerName(name: string) {
    this.playerName = name;
    localStorage.setItem(WEAKEST_PLAYER_NAME, JSON.stringify(this.playerName));
  }

  nextState(fromState: string | null = null) {
    this.execute('next', { from: fromState });
  }

  answerCorrect(isCorrect: boolean) {
    this.execute('answer', { correct: Boolean(isCorrect) });
  }

  saveBank() {
    this.execute('bank', {});
  }

  vote(weakestName: string) {
    this.execute('vote', {
      voter: this.playerName,
      weakest: weakestName,
    });
  }

  selectFinalAnswerer(playerName: string) {
    this.execute('final_answerer', { answerer: playerName });
  }

  hasPlayerName(): boolean {
    return Boolean(this.playerName);
  }
}
