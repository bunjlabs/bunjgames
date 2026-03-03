import React from "react";
import {useNavigate} from "react-router-dom";

import {Loading, useAuth, useGame} from "common/Essentials";
import {PlayerAuth} from "common/Auth";
import {GameClient, Content, Header, ExitButton, TextContent, BigButtonContent} from "common/Client";
import {FEUD_API} from "../index";


const stateContent = (game) => {
    const buttonActive = !game.answerer;

    const onButton = () => buttonActive && FEUD_API.button_click(FEUD_API.playerId);

    switch (game.state) {
        case "button":
            return <BigButtonContent active={buttonActive} onClick={onButton} />
        case "end":
            return <TextContent>Game over</TextContent>;
        default:
            return <TextContent>Friends Feud</TextContent>
    }
};

const FeudClient = () => {
    const game = useGame(FEUD_API);
    const [connected, setConnected] = useAuth(FEUD_API);
    const navigate = useNavigate();

    if (!connected) return <PlayerAuth api={FEUD_API} setConnected={setConnected}/>;
    if (!game) return <Loading/>;

    const onLogout = () => {
        FEUD_API.logout();
        navigate("/");
    };

    return <GameClient>
        <Header><ExitButton onClick={onLogout}/></Header>
        <Content>{stateContent(game)}</Content>
    </GameClient>
}

export default FeudClient;
