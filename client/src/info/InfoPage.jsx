import React from "react";
import { Link } from "react-router-dom";
import styles from "info/InfoPage.module.scss";
import classNames from "classnames";

const MainPage = () => {
    return <div className={styles.body}>
        <div className={styles.header}>
            <div className={styles.title}>Bunjgames</div>
            <div className={classNames(styles.textRight, styles.title)}><Link to={'/admin'}>Admin panel</Link></div>
        </div>
        <div className={styles.category}>
            <div className={styles.subtitle}><Link to={'/about'}>About page</Link></div>
        </div>
        <div className={styles.category}>
            <div className={styles.subtitle}>Whirligig</div>
            <div>Throughout the game, a team of six (recommended) experts attempts to answer questions sent in by viewers.
                For each question, the time limit is one minute. The questions require a combination of skills such as logical thinking,
                intuition, insight, etc. to find the correct answer.
                The team of experts earns points if they manage to get the correct answer.</div>
        </div>
        <div className={styles.category}>
            <div className={styles.subtitle}><Link to={'/jeopardy/client'}>Jeopardy</Link></div>
            <div>Three (recommended) contestants each take their place behind a lectern.
                The contestants compete in a quiz game comprising two or three rounds and Final round.
                The material for the clues covers a wide variety of topics.
                Category titles often feature puns, wordplay, or shared themes, and the host regularly reminds
                contestants of topics or place emphasis on category themes before the start of the round.</div>
        </div>
        <div className={styles.category}>
            <div className={styles.subtitle}><Link to={'/weakest/client'}>The Weakest</Link></div>
            <div>The format features 3-7 (recommended) contestants, who take turns answering general knowledge questions.
                The objective of every round is to create a chain of nine correct answers in a row and earn an increasing
                amount of money within a time limit.
                One wrong answer breaks the chain and loses any money earned within that particular chain.
                However, before their question is asked (but after their name is called), a contestant can choose
                to bank the current amount of money earned in any chain to make it safe, after which the chain starts afresh.
                A contestant's decision not to bank, in anticipation being able to correctly answer the upcoming question
                allows the money to grow, as each successive correct answer earns proportionally more money.</div>
        </div>
        <div className={styles.category}>
            <div className={styles.subtitle}><Link to={'/feud/client'}>Friends Feud</Link></div>
            <div>The team with control of the question then tries to win the round by guessing all of the remaining concealed answers,
                with each member giving one answer in sequence. Giving an answer not on the board, or failing to respond within the allotted time,
                earns one strike. If the team earns three strikes, their opponents are given one chance to "steal" the points for the round
                by guessing any remaining concealed answer; failing to do so awards the points back to the family that originally had control.
                If the opponents are given the opportunity to "steal" the points, then only their team's captain is required to answer the question.
                Any remaining concealed answers on the board that were not guessed are then revealed.</div>
        </div>
    </div>
}

const AdminPage = () => {
    return <div className={styles.body}>
        <div className={styles.header}>
            <div className={styles.title}>Bunjgames</div>
            <div className={classNames(styles.textRight, styles.title)}><Link to={'/'}>Home</Link></div>
        </div>
        <div className={styles.category}>
            <div className={styles.subtitle}>Whirligig:</div>
            <div><Link to={'/whirligig/admin'}>Admin panel</Link></div>
            <div><Link to={'/whirligig/view'}>View</Link></div>
        </div>
        <div className={styles.category}>
            <div className={styles.subtitle}>Jeopardy:</div>
            <div><Link to={'/jeopardy/admin'}>Admin panel</Link></div>
            <div><Link to={'/jeopardy/view'}>View</Link></div>
        </div>
        <div className={styles.category}>
            <div className={styles.subtitle}>The Weakest:</div>
            <div><Link to={'/weakest/admin'}>Admin panel</Link></div>
            <div><Link to={'/weakest/view'}>View</Link></div>
        </div>
        <div className={styles.category}>
            <div className={styles.subtitle}>Friends Feud:</div>
            <div><Link to={'/feud/admin'}>Admin panel</Link></div>
            <div><Link to={'/feud/view'}>View</Link></div>
        </div>
    </div>
}


const AboutPage = () => {
    return <div className={styles.body}>
        <div className={styles.header}>
            <div className={styles.title}>Bunjgames</div>
            <div className={classNames(styles.textRight, styles.title)}><Link to={'/'}>Home</Link></div>
        </div>

        <div className={styles.category}>
            <div className={styles.title}>How to play:</div>
            <div className={styles.marginBottom}/>

            <div>You should create the game using one of the selected game files or by creating your own game file.</div>
            <div className={styles.marginBottom}/>
            <div>Create new game using admin panel (you will be using it for the rest of the game) and then open it in
                view panel with game token (top left of the Admin panel).
                Preferably this should be a big enough screen for all of the game participants to view.
                I'm personally use my laptop for admin panel and cast View panel to my TV with chromecast in another tab.</div>
            <div className={styles.marginBottom}/>
            <div>Most of the games (all of them except whirligig for now) require their players to join it using game client.
                They can use game token or QR code available at first screen of View panel. Player name should be unique.
                If player exits the game by accident, he can reenter the game using his name and game token at any time.</div>
        </div>

        <div className={styles.category}>
            <div className={styles.title}>Where to find game packs:</div>
            <div className={styles.marginBottom}/>

            <div>
                <a href={'https://drive.google.com/drive/folders/1a4MoR8FusJCEePqR1SOxratkvchsgtRX?usp=sharing'}>
                    Game packs by bunjdo
                </a> (russian)
            </div>
            <div>You can also find game pack templates here. You can create your game using template an zip it.</div>
            <div className={styles.marginBottom}/>
            <div className={styles.subtitle}>Jeopardy:</div>
            <div className={styles.marginBottom}/>
            <div>
                <a href={'https://vladimirkhil.com/si/storage'}>
                    Official Jeopardy game packs
                </a> (russian)
            </div>
            <div>
                <a href={'https://vk.com/topic-135725718_34975471'}>
                    Unofficial Jeopardy game packs
                </a> (russian)
            </div>
        </div>

        <div className={styles.category}>
            <div className={styles.title}>Whirligig game file specification:</div>
            <div className={styles.marginBottom}/>

            <div className={styles.subtitle}>Zip archive file with structure:</div>
            <div className={styles.tab}>content.xml</div>
            <div className={styles.tab}>assets/  - images, audio and video folder</div>
            <div className={styles.marginBottom}/>

            <div className={styles.subtitle}>content.xml structure:</div>
            <div className={styles.marginBottom}/>
            <div className={classNames(styles.tab, styles.preWrap)}>{'<?xml version="1.0" encoding="utf-8"?>\n' +
            '<!DOCTYPE game>\n' +
            '<game>\n' +
            '    <items>  <!-- 13 items -->\n' +
            '        <item>\n' +
            '            <number>1</number>  <!-- integer -->\n' +
            '            <name>1</name>  <!-- string -->\n' +
            '            <description>question</description>  <!-- string -->\n' +
            '            <type>standard</type>  <!-- standard, blitz, superblitz -->\n' +
            '            <questions> <!-- 1 for standard, 3 for blitz and superblitz -->\n' +
            '                <question>\n' +
            '                    <description>question</description>  <!-- string -->\n' +
            '                    <text></text>  <!-- string, optional -->\n' +
            '                    <image></image>  <!-- string, optional -->\n' +
            '                    <audio></audio>  <!-- string, optional -->\n' +
            '                    <video></video>  <!-- string, optional -->\n' +
            '                    <answer>\n' +
            '                        <description>answer</description>  <!-- string -->\n' +
            '                        <text></text>  <!-- string, optional -->\n' +
            '                        <image></image>  <!-- string, optional -->\n' +
            '                        <audio></audio>  <!-- string, optional -->\n' +
            '                        <video></video>  <!-- string, optional -->\n' +
            '                    </answer>\n' +
            '                </question>\n' +
            '                ...  <!-- 1 item for standard question, 3 for blitz and superblitz -->\n' +
            '            </questions>\n' +
            '        </item>\n' +
            '   </items>\n' +
            '   ...\n' +
            '</game>'}</div>
        </div>

        <div className={styles.category}>
            <div className={styles.title}>Jeopardy game file specification:</div>
            <div className={styles.marginBottom}/>

            <div>
                <a href={'https://vladimirkhil.com/si/siquester'}>Jeopardy game packs editor (russian only)</a>
            </div>
            <div className={styles.marginBottom}/>
            <div>Coming soon...</div>
        </div>

        <div className={styles.category}>
            <div className={styles.title}>The Weakest game file specification:</div>
            <div className={styles.marginBottom}/>

            <div>Text (XML) file with following structure:</div>
            <div className={styles.marginBottom}/>
            <div className={classNames(styles.tab, styles.preWrap)}>{'<?xml version="1.0" encoding="UTF-8"?>\n' +
                '<!DOCTYPE game>\n' +
                '<game>\n' +
                '   <questions> <!-- Should contain a lot of questions, recommended amount is 100 - 200 -->\n' +
                '      <question>\n' +
                '         <question>question</question> <!-- string -->\n' +
                '         <answer>answer</answer> <!-- string -->\n' +
                '      </question>\n' +
                '      ...\n' +
                '   </questions>\n' +
                '   <final_questions> <!-- Minimum 10 questions, recommended 20 -->\n' +
                '      <question>\n' +
                '         <question>question</question> <!-- string -->\n' +
                '         <answer>answer</answer> <!-- string -->\n' +
                '      </question>\n' +
                '      ...\n' +
                '   </final_questions>\n' +
                '   <score_multiplier>1</score_multiplier> <!-- integer, determines the score multiplyer -->\n' +
                '</game>'}</div>
        </div>

        <div className={styles.category}>
            <div className={styles.title}>Friends Feud game file specification:</div>
            <div className={styles.marginBottom}/>

            <div>Coming soon...</div>
        </div>

        <div className={styles.category}>
            <div className={styles.title}>Friends Feud game file specification:</div>
            <div className={styles.marginBottom}/>

            <div>Coming soon...</div>
        </div>

        <div className={styles.category}>
            <div className={styles.title}>Assets specification:</div>
            <div className={styles.marginBottom}/>
            <div>You can place your assets anywhere at assets folder.</div>
            <div>Nested folders (assets/image, assets/audio, etc.) are optional.</div>
            <div>Leading slash (/) is mandatory.</div>
            <div>Assets encoding is not limited, but your target browser must be able to use it.</div>
            <div>Example:</div>
            <div className={styles.tab}>/image/1a.png</div>
            <div className={styles.tab}>/audio/music.mp3</div>
        </div>
    </div>
}

export {
    MainPage,
    AdminPage,
    AboutPage
}
