import React from 'react';
import { Link } from 'react-router-dom';

const linkStyle: React.CSSProperties = { color: 'var(--text)' };

const PageShell: React.FC<React.PropsWithChildren<{ rightLink: { to: string; label: string } }>> = ({ rightLink, children }) => (
  <div className="no-scrollbar" style={{ height: '100%', overflowY: 'scroll', scrollbarWidth: 'none', backgroundColor: 'var(--bg-base)', color: 'var(--text)' }}>
    <div style={{ backgroundColor: 'var(--bg-dark)', padding: 16, display: 'flex', justifyContent: 'space-between' }}>
      <div style={{ fontSize: 24 }}>Bunjgames</div>
      <div style={{ fontSize: 24, textAlign: 'right' }}><Link to={rightLink.to} style={linkStyle}>{rightLink.label}</Link></div>
    </div>
    {children}
  </div>
);

export const MainPage: React.FC = () => (
  <PageShell rightLink={{ to: '/admin', label: 'Admin panel' }}>
    <div style={{ padding: 16 }}>
      <div style={{ fontSize: 22 }}><Link to="/about" style={linkStyle}>About page</Link></div>
    </div>
    <div style={{ padding: 16 }}>
      <div style={{ fontSize: 22 }}>Whirligig</div>
      <div>Throughout the game, a team of six (recommended) experts attempts to answer questions sent in by viewers.
        For each question, the time limit is one minute. The questions require a combination of skills such as logical thinking,
        intuition, insight, etc. to find the correct answer. The team of experts earns points if they manage to get the correct answer.</div>
    </div>
    <div style={{ padding: 16 }}>
      <div style={{ fontSize: 22 }}><Link to="/jeopardy/client" style={linkStyle}>Jeopardy</Link></div>
      <div>Three (recommended) contestants each take their place behind a lectern.
        The contestants compete in a quiz game comprising two or three rounds and Final round.
        The material for the clues covers a wide variety of topics.</div>
    </div>
    <div style={{ padding: 16 }}>
      <div style={{ fontSize: 22 }}><Link to="/weakest/client" style={linkStyle}>The Weakest</Link></div>
      <div>The format features 3-7 (recommended) contestants, who take turns answering general knowledge questions.
        The objective of every round is to create a chain of nine correct answers in a row and earn an increasing
        amount of money within a time limit.</div>
    </div>
    <div style={{ padding: 16 }}>
      <div style={{ fontSize: 22 }}><Link to="/feud/client" style={linkStyle}>Friends Feud</Link></div>
      <div>The team with control of the question then tries to win the round by guessing all of the remaining concealed answers,
        with each member giving one answer in sequence.</div>
    </div>
  </PageShell>
);

export const AdminPage: React.FC = () => (
  <PageShell rightLink={{ to: '/', label: 'Home' }}>
    {[
      { name: 'Whirligig', admin: '/whirligig/admin', view: '/whirligig/view' },
      { name: 'Jeopardy', admin: '/jeopardy/admin', view: '/jeopardy/view' },
      { name: 'The Weakest', admin: '/weakest/admin', view: '/weakest/view' },
      { name: 'Friends Feud', admin: '/feud/admin', view: '/feud/view' },
    ].map((g) => (
      <div key={g.name} style={{ padding: 16 }}>
        <div style={{ fontSize: 22 }}>{g.name}:</div>
        <div><Link to={g.admin} style={linkStyle}>Admin panel</Link></div>
        <div><Link to={g.view} style={linkStyle}>View</Link></div>
      </div>
    ))}
  </PageShell>
);

export const AboutPage: React.FC = () => (
  <PageShell rightLink={{ to: '/', label: 'Home' }}>
    <div style={{ padding: 16 }}>
      <div style={{ fontSize: 24 }}>How to play:</div>
      <div style={{ marginBottom: 10 }} />
      <div>You should create the game using one of the selected game files or by creating your own game file.</div>
      <div style={{ marginBottom: 10 }} />
      <div>Create new game using admin panel (you will be using it for the rest of the game) and then open it in
        view panel with game token (top left of the Admin panel).
        Preferably this should be a big enough screen for all of the game participants to view.</div>
      <div style={{ marginBottom: 10 }} />
      <div>Most of the games (all of them except whirligig for now) require their players to join it using game client.
        They can use game token or QR code available at first screen of View panel. Player name should be unique.</div>
    </div>

    <div style={{ padding: 16 }}>
      <div style={{ fontSize: 24 }}>Where to find game packs:</div>
      <div style={{ marginBottom: 10 }} />
      <div><a href="https://drive.google.com/drive/folders/1a4MoR8FusJCEePqR1SOxratkvchsgtRX?usp=sharing" style={linkStyle}>Game packs by bunjdo</a> (russian)</div>
      <div>You can also find game pack templates here.</div>
      <div style={{ marginBottom: 10 }} />
      <div style={{ fontSize: 22 }}>Jeopardy:</div>
      <div style={{ marginBottom: 10 }} />
      <div><a href="https://vladimirkhil.com/si/storage" style={linkStyle}>Official Jeopardy game packs</a> (russian)</div>
      <div><a href="https://vk.com/topic-135725718_34975471" style={linkStyle}>Unofficial Jeopardy game packs</a> (russian)</div>
    </div>

    <div style={{ padding: 16 }}>
      <div style={{ fontSize: 24 }}>Whirligig game file specification:</div>
      <div style={{ marginBottom: 10 }} />
      <div style={{ fontSize: 22 }}>Zip archive file with structure:</div>
      <div style={{ paddingLeft: 40 }}>content.xml</div>
      <div style={{ paddingLeft: 40 }}>assets/ - images, audio and video folder</div>
      <div style={{ marginBottom: 10 }} />
      <div style={{ fontSize: 22 }}>content.xml structure:</div>
      <div style={{ marginBottom: 10 }} />
      <div style={{ paddingLeft: 40, whiteSpace: 'pre-wrap' }}>{'<?xml version="1.0" encoding="utf-8"?>\n<!DOCTYPE game>\n<game>\n    <items>  <!-- 13 items -->\n        <item>\n            <number>1</number>\n            <name>1</name>\n            <description>question</description>\n            <type>standard</type>\n            <questions>\n                <question>\n                    <description>question</description>\n                    <text></text>\n                    <image></image>\n                    <audio></audio>\n                    <video></video>\n                    <answer>\n                        <description>answer</description>\n                        <text></text>\n                        <image></image>\n                        <audio></audio>\n                        <video></video>\n                    </answer>\n                </question>\n            </questions>\n        </item>\n   </items>\n</game>'}</div>
    </div>

    <div style={{ padding: 16 }}>
      <div style={{ fontSize: 24 }}>Jeopardy game file specification:</div>
      <div style={{ marginBottom: 10 }} />
      <div><a href="https://vladimirkhil.com/si/siquester" style={linkStyle}>Jeopardy game packs editor (russian only)</a></div>
      <div style={{ marginBottom: 10 }} />
      <div>Coming soon...</div>
    </div>

    <div style={{ padding: 16 }}>
      <div style={{ fontSize: 24 }}>The Weakest game file specification:</div>
      <div style={{ marginBottom: 10 }} />
      <div>Text (XML) file with following structure:</div>
      <div style={{ marginBottom: 10 }} />
      <div style={{ paddingLeft: 40, whiteSpace: 'pre-wrap' }}>{'<?xml version="1.0" encoding="UTF-8"?>\n<!DOCTYPE game>\n<game>\n   <questions>\n      <question>\n         <question>question</question>\n         <answer>answer</answer>\n      </question>\n   </questions>\n   <final_questions>\n      <question>\n         <question>question</question>\n         <answer>answer</answer>\n      </question>\n   </final_questions>\n   <score_multiplier>1</score_multiplier>\n</game>'}</div>
    </div>

    <div style={{ padding: 16 }}>
      <div style={{ fontSize: 24 }}>Friends Feud game file specification:</div>
      <div style={{ marginBottom: 10 }} />
      <div>Coming soon...</div>
    </div>

    <div style={{ padding: 16 }}>
      <div style={{ fontSize: 24 }}>Assets specification:</div>
      <div style={{ marginBottom: 10 }} />
      <div>You can place your assets anywhere at assets folder.</div>
      <div>Nested folders (assets/image, assets/audio, etc.) are optional.</div>
      <div>Leading slash (/) is mandatory.</div>
      <div>Assets encoding is not limited, but your target browser must be able to use it.</div>
      <div>Example:</div>
      <div style={{ paddingLeft: 40 }}>/image/1a.png</div>
      <div style={{ paddingLeft: 40 }}>/audio/music.mp3</div>
    </div>
  </PageShell>
);
