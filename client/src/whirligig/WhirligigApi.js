import GameApi from "../gameApi.js";

const WHIRLIGIG_TOKEN = "WHIRLIGIG_TOKEN";

export default class WhirligigApi extends GameApi {
    constructor(apiEndpoint, wsEndpoint) {
        super(apiEndpoint, wsEndpoint, WHIRLIGIG_TOKEN);
    }

    createGame(inputFile) {
        const formData = new FormData();
        formData.append("game", inputFile.files[0]);

        return this.axios.post('create', formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        }).then(result => {
            this.saveToken(result.data.token);
            return result.data;
        });
    }

    calcTime() {
        const serverTime = this.game.timer_time;
        const serverPausedTime = this.game.timer_paused_time;
        const isPaused = this.game.timer_paused;
        const now = Date.now();

        let time;
        if (isPaused) {
            time = Math.round((serverTime - serverPausedTime) / 1000);
        } else {
            time = Math.round((serverTime - now) / 1000);
        }
        return time;
    }

    score(connoisseurs_score, viewers_score) {
        this.execute("change_score", {connoisseurs_score, viewers_score})
    }

    timer(paused) {
        this.execute("change_timer", {paused})
    }

    answerCorrect(isCorrect) {
        this.execute("answer_correct", {is_correct: Boolean(isCorrect)})
    }

    nextState(fromState=null) {
        this.execute("next_state", {"from_state": fromState})
    }

    extra_time() {
        this.execute("extra_time", {})
    }
}
