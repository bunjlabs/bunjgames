import GameApi from './GameApi';

const JEOPARDY_TOKEN = 'JEOPARDY_TOKEN';
const JEOPARDY_PLAYER_NAME = 'JEOPARDY_PLAYER_NAME';

export default class JeopardyApi extends GameApi {
  playerName: string | null = null;

  constructor(apiEndpoint: string, wsEndpoint: string) {
    super(apiEndpoint, wsEndpoint, JEOPARDY_TOKEN);
    this.loadPlayerName();
  }

  createGame(inputFile: HTMLInputElement) {
    const formData = new FormData();
    formData.append('game', 'jeopardy');
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

  loadPlayerName(): string | null {
    try {
      this.playerName = JSON.parse(
        localStorage.getItem(JEOPARDY_PLAYER_NAME) ?? 'null',
      );
    } catch {
      this.playerName = null;
    }
    return this.playerName;
  }

  savePlayerName(name: string) {
    this.playerName = name;
    localStorage.setItem(JEOPARDY_PLAYER_NAME, JSON.stringify(this.playerName));
  }

  nextState(fromState: string | null = null) {
    this.execute('next', { from: fromState });
  }

  chooseQuestion(question: string) {
    this.execute('chooseQuestion', { question });
  }

  setAnswererAndBet(player: string, bet: number) {
    this.execute('setAnswererAndBet', { player, bet });
  }

  skipQuestion() {
    this.execute('skipQuestion', {});
  }

  buttonClick() {
    if (this.playerName) {
      this.execute('buttonClick', { player: this.playerName });
    }
  }

  answer(isRight: boolean) {
    this.execute('answer', { correct: isRight });
  }

  removeFinalTheme(theme: string) {
    this.execute('removeFinalTheme', { theme });
  }

  finalBet(bet: number) {
    this.execute('finalBet', { player: this.playerName, bet });
  }

  finalAnswer(answer: string) {
    this.execute('finalAnswer', { player: this.playerName, answer });
  }

  finalPlayerAnswer(isRight: boolean) {
    this.execute('finalPlayerAnswer', { correct: isRight });
  }

  setBalance(balanceList: number[]) {
    this.execute('setBalance', { balanceList });
  }

  setRound(round: number) {
    this.execute('setRound', { round });
  }

  hasPlayerName(): boolean {
    return Boolean(this.playerName);
  }
}
