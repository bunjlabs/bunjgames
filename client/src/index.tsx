import {createRoot} from "react-dom/client";
import React from "react";
import {Howler} from 'howler';
import 'index.css';
import {BrowserRouter, Routes} from 'react-router-dom'
import {Route} from "react-router";
import {MainPage, AdminPage, AboutPage} from "./info/InfoPage";
import WhirligigAdmin from "./whirligig/Admin.jsx";
import WhirligigView from "./whirligig/View.jsx";
import WhirligigApi from "./whirligig/WhirligigApi.js";
import JeopardyApi from "./jeopardy/JeopardyApi.js";
import WeakestApi from "./weakest/WeakestApi.js";
import FeudApi from "./feud/FeudApi.js";
import JeopardyAdmin from "./jeopardy/Admin.jsx";
import JeopardyView from "./jeopardy/View.jsx";
import JeopardyClient from "./jeopardy/Client.jsx";
import WeakestAdmin from "./weakest/Admin.jsx";
import WeakestView from "./weakest/View.jsx";
import WeakestClient from "./weakest/Client.jsx";
import FeudAdmin from "./feud/Admin.jsx";
import FeudView from "./feud/View.jsx";
import FeudClient from "./feud/Client.jsx";

require("./polyfils.js");

export const BunjGamesConfig = {
    MEDIA: "/media/",

    COMMON_API_ENDPOINT: "/api/common/",

    WHIRLIGIG_API_ENDPOINT: "/api/whirligig/",
    WHIRLIGIG_WS_ENDPOINT: "ws://localhost:8000/ws/whirligig/",

    JEOPARDY_API_ENDPOINT: "/api/jeopardy/",
    JEOPARDY_WS_ENDPOINT: "/ws/jeopardy/",

    WEAKEST_API_ENDPOINT: "/api/weakest/",
    WEAKEST_WS_ENDPOINT: "ws://localhost:800/ws/weakest/",

    FEUD_API_ENDPOINT: "/api/feud/",
    FEUD_WS_ENDPOINT: "/ws/feud/",
};

export const WHIRLIGIG_API = new WhirligigApi(BunjGamesConfig.WHIRLIGIG_API_ENDPOINT, BunjGamesConfig.WHIRLIGIG_WS_ENDPOINT);
export const JEOPARDY_API = new JeopardyApi(BunjGamesConfig.JEOPARDY_API_ENDPOINT, BunjGamesConfig.JEOPARDY_WS_ENDPOINT);
export const WEAKEST_API = new WeakestApi(BunjGamesConfig.WEAKEST_API_ENDPOINT, BunjGamesConfig.WEAKEST_WS_ENDPOINT);
export const FEUD_API = new FeudApi(BunjGamesConfig.FEUD_API_ENDPOINT, BunjGamesConfig.FEUD_WS_ENDPOINT);

Howler.volume(0.5);

const App = () => {
    return <React.StrictMode>
        <BrowserRouter>
            <Routes>
                <Route index element={<MainPage />}/>
                <Route path="/admin" element={<AdminPage />}/>
                <Route path="/about" element={<AboutPage />}/>

                <Route path="/whirligig/admin" element={<WhirligigAdmin />}/>
                <Route path="/whirligig/view" element={<WhirligigView />}/>

                <Route path="/jeopardy/admin" element={<JeopardyAdmin />}/>
                <Route path="/jeopardy/view" element={<JeopardyView />}/>
                <Route path="/jeopardy/client" element={<JeopardyClient />}/>

                <Route path="/weakest/admin" element={<WeakestAdmin />}/>
                <Route path="/weakest/view" element={<WeakestView />}/>
                <Route path="/weakest/client" element={<WeakestClient />}/>

                <Route path="/feud/admin" element={<FeudAdmin />}/>
                <Route path="/feud/view" element={<FeudView />}/>
                <Route path="/feud/client" element={<FeudClient />}/>
            </Routes>
        </BrowserRouter>
    </React.StrictMode>
};

const root = document.getElementById("root") as HTMLElement;

createRoot(root).render(<App />);
