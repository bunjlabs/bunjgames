import GameApi from "gameApi.js";

const FEUD_TOKEN = "FEUD_TOKEN";
const FEUD_PLAYER_ID = "FEUD_PLAYER_ID";

export default class FeudApi extends GameApi {
    constructor(apiEndpoint, wsEndpoint) {
        super(apiEndpoint, wsEndpoint, FEUD_TOKEN);
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
        const formData = new FormData();
        formData.append("token", token);
        formData.append("name", name);

        return this.axios.post('players/register', formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
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
            this.playerId = JSON.parse(localStorage.getItem(FEUD_PLAYER_ID));
        } catch (e) {
            this.playerId = null;
        }
        return this.playerId;
    }

    savePlayerId(playerId) {
        this.playerId = playerId;
        localStorage.setItem(FEUD_PLAYER_ID, JSON.stringify(this.playerId));
    }

    next_state(from_state=null) {
        this.execute("next_state", {from_state})
    }

    button_click(player_id) {
        this.execute("button_click", {player_id})
    }

    set_answerer(player_id) {
        this.execute("set_answerer", {player_id})
    }

    answer(is_correct, answer_id) {
        this.execute("answer", {is_correct: Boolean(is_correct), answer_id: answer_id})
    }

    hasPlayerId() {
        return Boolean(this.playerId);
    }
}
