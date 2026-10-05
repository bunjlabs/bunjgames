import GameApi from './GameApi';

const DEAL_TOKEN = 'DEAL_TOKEN';

export default class DealApi extends GameApi {
  constructor(apiEndpoint: string, wsEndpoint: string) {
    super(apiEndpoint, wsEndpoint, DEAL_TOKEN);
  }

  createGame(inputFile: HTMLInputElement) {
    const formData = new FormData();
    formData.append('game', 'deal');
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

  nextState(fromState: string | null = null) {
    this.execute('next', { from: fromState });
  }

  selectCase(index: number) {
    this.execute('selectCase', { index });
  }

  openCase(index: number) {
    this.execute('openCase', { index });
  }

  revealCase(index: number) {
    this.execute('revealCase', { index });
  }

  chooseMoney() {
    this.execute('chooseMoney', {});
  }

  shuffle() {
    this.execute('shuffle', {});
  }

  keep() {
    this.execute('keep', {});
  }

  switchCase() {
    this.execute('switch', {});
  }

  keepGift() {
    this.execute('keepGift', {});
  }

  spinWheel() {
    this.execute('spinWheel', {});
  }

  setCheat(index: number) {
    this.execute('setCheat', { index });
  }

  spin() {
    this.execute('spin', {});
  }
}
