import { createRoot } from 'react-dom/client';
import React from 'react';
import { Howler } from 'howler';
import 'index.css';
import 'services/polyfills';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { MainPage, AdminPage, AboutPage } from './info/InfoPage';
import WhirligigAdmin from './whirligig/Admin';
import WhirligigView from './whirligig/View';
import JeopardyAdmin from './jeopardy/Admin';
import JeopardyView from './jeopardy/View';
import JeopardyClient from './jeopardy/Client';
import WeakestAdmin from './weakest/Admin';
import WeakestView from './weakest/View';
import WeakestClient from './weakest/Client';
import FeudAdmin from './feud/Admin';
import FeudView from './feud/View';
import FeudClient from './feud/Client';

Howler.volume(0.5);

const App = () => (
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route index element={<MainPage />} />
        <Route path="/admin" element={<AdminPage />} />
        <Route path="/about" element={<AboutPage />} />

        <Route path="/whirligig/admin" element={<WhirligigAdmin />} />
        <Route path="/whirligig/view" element={<WhirligigView />} />

        <Route path="/jeopardy/admin" element={<JeopardyAdmin />} />
        <Route path="/jeopardy/view" element={<JeopardyView />} />
        <Route path="/jeopardy/client" element={<JeopardyClient />} />

        <Route path="/weakest/admin" element={<WeakestAdmin />} />
        <Route path="/weakest/view" element={<WeakestView />} />
        <Route path="/weakest/client" element={<WeakestClient />} />

        <Route path="/feud/admin" element={<FeudAdmin />} />
        <Route path="/feud/view" element={<FeudView />} />
        <Route path="/feud/client" element={<FeudClient />} />
      </Routes>
    </BrowserRouter>
  </React.StrictMode>
);

createRoot(document.getElementById('root') as HTMLElement).render(<App />);
