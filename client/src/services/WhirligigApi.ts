import GameApi from './GameApi';

const WHIRLIGIG_TOKEN = 'WHIRLIGIG_TOKEN';

export default class WhirligigApi extends GameApi {
  constructor(apiEndpoint: string, wsEndpoint: string) {
    super(apiEndpoint, wsEndpoint, WHIRLIGIG_TOKEN);
  }

  createGame(inputFile: HTMLInputElement) {
    const formData = new FormData();
    formData.append('game', 'whirligig');
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

  getTime(from: number): number {
    const serverTime = this.game!.timer.time;
    const paused = this.game!.timer.paused;

    if (paused) return Math.max(Math.floor(serverTime / 1000), 0);

    const now = Date.now();
    return Math.max(Math.floor((serverTime - (now - from)) / 1000), 0);
  }

  score(connoisseurs: number, viewers: number) {
    this.execute('score', { connoisseurs, viewers });
  }

  timer(paused: boolean) {
    this.execute('timer', { paused });
  }

  answerCorrect(isCorrect: boolean) {
    this.execute('answer', { correct: Boolean(isCorrect) });
  }

  nextState(fromState: string | null = null) {
    this.execute('next', { from: fromState });
  }

  extraTime() {
    this.execute('extraTime', {});
  }
}
