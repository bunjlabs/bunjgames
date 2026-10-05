import React from 'react';

const itemStyle: React.CSSProperties = {
  padding: '8px 12px',
  backgroundColor: 'var(--bg-dark)',
  color: 'var(--text)',
  fontSize: 20,
  textAlign: 'center',
  whiteSpace: 'nowrap',
  overflow: 'hidden',
  textOverflow: 'ellipsis',
};

export const PresentsList: React.FC<{
  game: any;
  showDescription: boolean;
  onClick?: (index: number) => void;
  onMoney?: () => void;
}> = ({ game, showDescription, onClick, onMoney }) => {
  const presents = game.presents as any[];

  const items = presents
    .map((p: any, i: number) => ({ p, i }))
    .filter(({ p }) => !p.given)
    .map(({ p, i }) => {
      const struck = p.opened;
      return (
        <div
          key={i}
          className={onClick ? 'clickable' : ''}
          style={{
            ...itemStyle,
            marginBottom: 8,
            textDecoration: struck ? 'line-through' : 'none',
            cursor: onClick ? 'pointer' : 'default',
          }}
          onClick={() => onClick?.(i)}
        >
          <div>{p.content}</div>
          {showDescription && p.description && (
            <div style={{ fontSize: 14, fontWeight: 'normal', color: 'var(--text-gray)' }}>{p.description}</div>
          )}
        </div>
      );
    });

  return (
    <div className="no-scrollbar" style={{ width: '100%', flex: '0 1 auto', minHeight: 0 }}>
      {items}
      {onMoney && (
        <>
          <div style={{ borderTop: '2px solid var(--bg-button)', margin: '4px 0' }} />
          <div
            className="clickable"
            style={{ ...itemStyle, cursor: 'pointer', fontWeight: 'bold' }}
            onClick={onMoney}
          >
            Money
          </div>
        </>
      )}
    </div>
  );
};
