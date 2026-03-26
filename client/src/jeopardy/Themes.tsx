import React from 'react';
import { FaCat } from 'react-icons/fa';
import { RiAuctionFill } from 'react-icons/ri';

const themeTextStyle: React.CSSProperties = {
  textTransform: 'uppercase', color: 'var(--text)', margin: '0 10px', fontSize: 24,
  textAlign: 'center', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
};

const questionValueStyle: React.CSSProperties = {
  textTransform: 'uppercase', color: 'var(--text-value)', textShadow: '1px 2px 3px #000',
  fontSize: 38, fontWeight: 'bold', textAlign: 'center',
  overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
};

const cellStyle: React.CSSProperties = {
  backgroundColor: 'var(--bg-dark)', display: 'flex', flexDirection: 'column', justifyContent: 'center',
};

const Theme: React.FC<{ theme: any; onSelect?: (name: string) => void; active?: boolean }> = ({
  theme, onSelect, active = false,
}) => (
  <div
    style={cellStyle}
    className={active ? 'clickable' : ''}
    onClick={() => active && !theme.isRemoved && onSelect?.(theme.name)}
    title={theme.name + '\n' + (theme.comment ?? '')}
  >
    <div style={themeTextStyle}>{!theme.isRemoved && theme.name}</div>
  </div>
);

const ThemeNameCell: React.FC<{ name: string }> = ({ name }) => (
  <div style={cellStyle}>
    <div style={themeTextStyle}>{name}</div>
  </div>
);

const Question: React.FC<{ question: any; themeName: string; onSelect?: (key: string) => void }> = ({
  question, themeName, onSelect,
}) => {
  const key = `${themeName}:${question.value}`;
  return (
    <div
      style={cellStyle}
      className={!question.isProcessed && onSelect ? 'clickable' : ''}
      onClick={() => !question.isProcessed && onSelect?.(key)}
    >
      <div style={questionValueStyle}>
        {!question.isProcessed && onSelect && question.type === 'bagcat' && <FaCat />}
        {!question.isProcessed && onSelect && question.type === 'auction' && <RiAuctionFill />}
        {!question.isProcessed && question.value}
      </div>
    </div>
  );
};

export const ThemesList: React.FC<{
  game: any; onSelect?: (name: string) => void; active?: boolean;
}> = ({ game, onSelect, active = false }) => {
  const themes = game.round?.themes ?? [];
  return (
    <div style={{ height: '100%', width: '100%', fontFamily: 'Verdana, sans-serif', padding: '10px 25%', display: 'grid', gridTemplateColumns: 'repeat(1, minmax(0, 1fr))', gridAutoRows: '1fr', gap: 10 }}>
      {themes.map((theme: any, i: number) => (
        <Theme key={i} theme={theme} onSelect={onSelect} active={active} />
      ))}
    </div>
  );
};

export const ThemesGrid: React.FC<{ game: any }> = ({ game }) => {
  const themeNames: string[] = game.themes ?? [];
  return (
    <div style={{ height: '100%', width: '100%', fontFamily: 'Verdana, sans-serif', padding: 10, display: 'grid', gridTemplateColumns: 'repeat(3, minmax(0, 1fr))', gridAutoRows: 'minmax(10px, auto)', gap: 10 }}>
      {themeNames
        .slice(0, themeNames.length - (themeNames.length % 3))
        .map((name: string, i: number) => <ThemeNameCell key={i} name={name} />)}
    </div>
  );
};

export const QuestionsGrid: React.FC<{
  game: any; onSelect?: (key: string) => void;
}> = ({ game, onSelect }) => {
  const themes = game.round?.themes ?? [];
  const maxQ = Math.max(...themes.map((t: any) => t.questions.length));
  const items: React.ReactNode[] = [];

  themes.forEach((theme: any, ti: number) => {
    items.push(<Theme key={ti} theme={theme} />);
    theme.questions.forEach((q: any, qi: number) => (
      items.push(<Question key={`${ti}_${qi}`} question={q} themeName={theme.name} onSelect={onSelect} />)
    ));
  });

  return (
    <div style={{ height: '100%', width: '100%', fontFamily: 'Verdana, sans-serif', padding: 10, display: 'grid', gap: 10, gridTemplateColumns: `minmax(0, 5fr) repeat(${maxQ}, minmax(0, 2fr))` }}>
      {items}
    </div>
  );
};
