import React from "react";
import {useNavigate} from "react-router-dom";

import {VerticalList, ListItem, Loading, useAuth, useGame} from "common/Essentials";
import {PlayerAuth} from "common/Auth";
import {GameClient, Content, Header, ExitButton, TextContent, BigButtonContent} from "common/Client";

import styles from "weakest/Client.module.scss";
import classNames from "classnames";
import {WEAKEST_API} from "../index";


const Players = ({game, player, onClick}) => {
    return <VerticalList className={styles.players}>
        {game.players.filter(p => !p.is_weak).map(p =>
            <ListItem
                key={p.id}
                className={classNames(
                    styles.player,
                    p.id === player.id && styles.self,
                    p.id === player.weak_id && styles.selected,
                    p.id !== player.id && styles.active
                )}
                onClick={() => p.id !== player.id && onClick(p.id)}
            >
                {p.name}
            </ListItem>
        )}
    </VerticalList>
}


const stateContent = (game) => {
    const player = game.players.find(p => p.id === WEAKEST_API.playerId);
    const buttonActive = game.answerer && game.answerer === WEAKEST_API.playerId;

    const onBank = () => buttonActive && WEAKEST_API.save_bank();
    const onPlayerSelect = (playerId) => WEAKEST_API.select_weakest(playerId);

    switch (game.state) {
        case "questions":
            return <BigButtonContent active={buttonActive} onClick={onBank}>Bank ({game.tmp_score})</BigButtonContent>
        case "weakest_choose":
            return <Players game={game} player={player} onClick={onPlayerSelect}/>
        case "end":
            return <TextContent>Game over</TextContent>;
        default:
            return <TextContent>The Weakest</TextContent>
    }
};

const WeakestClient = () => {
    const game = useGame(WEAKEST_API);
    const [connected, setConnected] = useAuth(WEAKEST_API);
    const navigate = useNavigate();

    if (!connected) return <PlayerAuth api={WEAKEST_API} setConnected={setConnected}/>;
    if (!game) return <Loading/>;

    const onLogout = () => {
        WEAKEST_API.logout();
        navigate("/");
    };

    return <GameClient>
        <Header><ExitButton onClick={onLogout}/></Header>
        <Content>{stateContent(game)}</Content>
    </GameClient>
}

export default WeakestClient;
