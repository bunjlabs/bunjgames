import React, {useEffect, useState} from "react";
import ReactPlayer from "react-player";
import styles from "./Essentials.module.scss";
import {ToastContainer} from "react-toastify";
import 'react-toastify/dist/ReactToastify.css';
import {Link} from "react-router-dom";
import {Howl} from "howler";
import classNames from "classnames";

const HowlWrapper = (src, loop = false, volume = 1.0) => {
    return new Howl({
        src: [src],
        loop: loop,
        volume: volume,
        preload: false
    });
}
const getMediaUrl = (game, url) => url.startsWith("/") ? `/media/${game.name}/${game.token}${url}` : url;

const ImagePlayer = ({game, url}) => (
    <img src={getMediaUrl(game, url)} alt="Missing"/>
);

const AudioPlayer = ({game, url, controls, playing}) => (
    <ReactPlayer controls={controls} playing={playing} url={getMediaUrl(game, url)} width="100%" height="100%"/>
);

const VideoPlayer = ({game, url, controls, playing}) => (
    <ReactPlayer controls={controls} playing={playing} url={getMediaUrl(game, url)} width="100%" height="100%"/>
);

const Loading = () => (
    <div className={styles.loading}>Loading...</div>
)

const Toast = () => (
    <ToastContainer
        position="top-right"
        autoClose={3000}
        hideProgressBar
        newestOnTop={false}
        closeOnClick
        rtl={false}
        pauseOnFocusLoss
        draggable
        pauseOnHover400px
    />
)

const Button = ({onClick, className, children}) => (
    <div className={classNames(styles.button, className)} onClick={onClick}>{children}</div>
);

const OvalButton = ({onClick, className, children}) => (
    <div className={classNames(styles.oval, styles.button, className)} onClick={onClick}>{children}</div>
);

const ButtonLink = ({to, className, children}) => (
    <Link className={classNames(styles.button, className)} to={to}>{children}</Link>
);

const Input = ({type, onChange, value, className}) => (
    <input className={classNames(className, styles.input)} type={type} onChange={onChange} value={value}/>
)

const VerticalList = ({className, children}) => (
    <div className={classNames(styles.verticalList, styles.list, className)}>
        {children}
    </div>
);

const HorizontalList = ({className, children}) => (
    <div className={classNames(styles.horizontalList, styles.list, className)}>
        {children}
    </div>
);

const ListItem = ({className, children, ...props}) => (
    <div className={classNames(styles.listItem, className)} {...props}>
        {children}
    </div>
);

const TwoLineListItem = ({className, children, ...props}) => (
    <div className={classNames(styles.twoLineListItem, styles.listItem, className)} {...props}>
        {children}
    </div>
);

const useGame = (api, onState=() => {}, onIntercom=() => {}) => {
    const [game, setGame] = useState();

    useEffect(() => {
        const gameId = api.getGameSubscriber().subscribe(setGame);
        const stateId = api.getStateSubscriber().subscribe(onState);
        const intercomId = api.getIntercomSubscriber().subscribe(onIntercom);
        return () => {
            api.getGameSubscriber().unsubscribe(gameId);
            api.getStateSubscriber().unsubscribe(stateId);
            api.getIntercomSubscriber().unsubscribe(intercomId);
        }
    }, [api, onIntercom, onState]);

    return game;
}

const useAuth = (api) => {
    const [connected, setConnected] = useState();
    useEffect(() => setConnected(api.isConnected()), [api]);
    return [connected, setConnected];
}


const useTimer = (api, onTimerEnd) => {
    const [time, setTime] = useState(api.calcTime());

    useEffect(() => {
        let timer;
        if (time <= 0) {
            if(onTimerEnd) onTimerEnd();
        } else if (time > 0) {
            timer = setInterval(() => {
                setTime(api.calcTime());
            }, 1000);
        }

        return () => timer && clearInterval(timer);
    });
    return time;
}

const calcStateName = (state) => {
    const value = state.replaceAll("_", " ");
    return value.charAt(0).toUpperCase() + value.slice(1);
}

export {
    HowlWrapper,
    ImagePlayer, AudioPlayer, VideoPlayer,
    Loading, Toast,
    Button, OvalButton, ButtonLink,
    Input,
    VerticalList, HorizontalList, ListItem, TwoLineListItem,
    useGame, useAuth, useTimer,
    calcStateName
}
