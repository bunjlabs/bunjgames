import React from "react";
import {Toast} from "common/Essentials";
import styles from "common/Client.module.scss";
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

const Header = ({children}) => (
    <div className={styles.header}>
        {children}
    </div>
);

const TextContent = ({children}) => (
    <div className={styles.text}>{children}</div>
)

const FormContent = ({children}) => (
    <div className={styles.form}>
        {children}
    </div>
)

const BigButtonContent = ({active, onClick, children}) => (
    <div className={classNames(styles.bigButton, active && styles.active)} onClick={onClick} onTouchStart={onClick}>
        {children}
    </div>
)

const Content = ({children}) => (
    <div className={styles.content}>
        {children}
    </div>
)

const GameClient = ({children}) => (
    <div className={styles.client}>
        {children}
        <Toast/>
    </div>
);

export {
    ExitButton,
    Header,
    TextContent,
    FormContent,
    BigButtonContent,
    Content,
    GameClient,
};
