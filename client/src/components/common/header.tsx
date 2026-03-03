import {ReactNode} from "react";

type HeaderProps = {
    children?: ReactNode;
};

const Header = ({children}: HeaderProps) => {
    return <div style={{
        background: "#f5f5f5",
        padding: "16px",
        display: "flex",
        justifyContent: "space-between"
    }}>
        {children}
    </div>
}

export {
    Header
}
