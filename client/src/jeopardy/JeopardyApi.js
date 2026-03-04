import GameApi from "gameApi.js";

const JEOPARDY_TOKEN = "JEOPARDY_TOKEN";
const JEOPARDY_PLAYER_ID = "JEOPARDY_PLAYER_ID";

export default class JeopardyApi extends GameApi {
    constructor(apiEndpoint, wsEndpoint) {
        super(apiEndpoint, wsEndpoint, JEOPARDY_TOKEN);
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

    loadPlayerId() {
        try {
            this.playerId = JSON.parse(localStorage.getItem(JEOPARDY_PLAYER_ID));
        } catch (e) {
            this.playerId = null;
        }
        return this.playerId;
    }

    savePlayerId(playerId) {
        this.playerId = playerId;
        localStorage.setItem(JEOPARDY_PLAYER_ID, JSON.stringify(this.playerId));
    }

    nextState(from_state=null) {
        this.execute("next_state", {from_state})
    }

    chooseQuestion(question_id) {
        this.execute("choose_question", {question_id})
    }

    set_answerer_and_bet(player_id, bet) {
        this.execute("set_answerer_and_bet", {player_id, bet})
    }

    skip_question() {
        this.execute("skip_question", {})
    }

    button_click() {
        if(this.playerId) {
            this.execute("button_click", {player_id: this.playerId})
        }
    }

    answer(is_right) {
        this.execute("answer", {is_right})
    }

    remove_final_theme(theme_id) {
        this.execute("remove_final_theme", {theme_id})
    }

    final_bet(bet) {
        this.execute("final_bet", {player_id: this.playerId, bet})
    }

    final_answer(answer) {
        this.execute("final_answer", {player_id: this.playerId, answer})
    }

    final_player_answer(is_right) {
        this.execute("final_player_answer", {is_right})
    }

    set_balance(balance_list) {
        this.execute("set_balance", {balance_list})
    }

    set_round(round) {
        this.execute("set_round", {round})
    }

    hasPlayerId() {
        return Boolean(this.playerId);
    }
}
