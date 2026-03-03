import React from "react";
import {QRCodeSVG} from "qrcode.react";
import styles from "common/View.module.scss";
import {Toast} from "common/Essentials";
import {FaTimesCircle} from "react-icons/fa";
import classNames from "classnames";

const ExitButton = ({onClick}) => (
    <button className={styles.exit} onClick={e => {
        if(window.confirm("Are you sure want to exit?")){
            onClick();
        } else {
            e.preventDefault();
        }
    }}><FaTimesCircle /></button>
)

const TextContent = ({className, children}) => (
    <div className={classNames(styles.text, className)}>
        <p>{children}</p>
    </div>
);

const generateClientUrl = (path) => {
    return window.location.protocol + '//' + window.location.host + path;
}

const QRCodeContent = ({className, children, value}) => (
    <div className={classNames(styles.text, className)}>
        <p>{children}</p>
        <QRCodeSVG className={styles.qr} size={2000} marginSize={4} bgColor={'#fff'} value={value} />
    </div>
);

const BlockContent = ({children}) => (
    <div className={styles.block}>
        {children}
    </div>
);

const Content = ({children}) => (
    <div className={styles.content}>
        {children}
    </div>
);

const GameView = ({children}) => (
    <div className={styles.view}>
        {children}
        <Toast/>
    </div>
);

export {
    ExitButton,
    TextContent,
    QRCodeContent,
    BlockContent,
    Content,
    GameView,
    generateClientUrl,
}
