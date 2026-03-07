import GameApi from "gameApi.js";

const WEAKEST_TOKEN = "WEAKEST_TOKEN";
const WEAKEST_PLAYER_ID = "WEAKEST_PLAYER_ID";

export default class WeakestApi extends GameApi {
    constructor(apiEndpoint, wsEndpoint) {
        super(apiEndpoint, wsEndpoint, WEAKEST_TOKEN);
        this.playerId = null;
        this.loadPlayerId();
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

    registerPlayer(token, name) {
        return this.axios.post('players/register', {
            token: token,
            name: name
        }).then(result => {
            this.saveToken(result.data.game.token);
            this.savePlayerId(result.data.player_id);
            return result.data;
        });
    }

    calcTime() {
        const now = Date.now();
        const time = Math.round((this.game.timer - now) / 1000);
        return Math.clamp(time, Number.MAX_VALUE,0);
    }

    loadPlayerId() {
        try {
            this.playerId = JSON.parse(localStorage.getItem(WEAKEST_PLAYER_ID));
        } catch (e) {
            this.playerId = null;
        }
        return this.playerId;
    }

    savePlayerId(playerId) {
        this.playerId = playerId;
        localStorage.setItem(WEAKEST_PLAYER_ID, JSON.stringify(this.playerId));
    }

    next_state(from_state=null) {
        this.execute("next_state", {from_state})
    }

    answer_correct(isCorrect) {
        this.execute("answer_correct", {is_correct: Boolean(isCorrect)})
    }

    save_bank() {
        this.execute("save_bank", {})
    }

    select_weakest(weakest_id) {
        this.execute("select_weakest", {player_id: this.playerId, weakest_id: weakest_id})
    }

    select_final_answerer(player_id) {
        this.execute("select_final_answerer", {player_id})
    }

    hasPlayerId() {
        return Boolean(this.playerId);
    }
}
