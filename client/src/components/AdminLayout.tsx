import React from 'react';
import { toast } from 'react-toastify';
import { Toast } from './UI';

export const AdminHeader: React.FC<React.PropsWithChildren<{
  gameName?: string;
  token: string;
  stateName: string;
}>> = ({ gameName, token, stateName, children }) => (
  <div style={{ height: 60, backgroundColor: 'var(--bg-dark)', color: 'var(--text)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0 32px' }}>
    <div style={{ fontSize: 22 }}>{gameName ?? 'Admin'}</div>
    <div
      style={{ fontSize: 22, cursor: 'pointer' }}
      onClick={() => { navigator.clipboard.writeText(token.toUpperCase()); toast('Token copied!'); }}
      title="Click to copy"
    >
      {token.toUpperCase()}
    </div>
    <div style={{ fontSize: 22, textAlign: 'center' }}>{stateName}</div>
    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>{children}</div>
  </div>
);

export const AdminContent: React.FC<React.PropsWithChildren<{
  rightPanel?: React.ReactNode;
}>> = ({ rightPanel, children }) => (
  <div style={{ flexGrow: 1, alignItems: 'stretch', color: 'var(--text)', display: 'flex', height: 0 }}>
    <div className="no-scrollbar" style={{ overflowY: 'scroll', scrollbarWidth: 'none', flexGrow: 1 }}>{children}</div>
    <div style={{ borderLeft: '8px solid var(--bg-dark)', width: '25%', minWidth: '25%', overflow: 'hidden', display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>{rightPanel}</div>
  </div>
);

export const BlockContent: React.FC<React.PropsWithChildren> = ({ children }) => {
  const items = React.Children.toArray(children);
  return (
    <div>
      {items.map((child, i) => (
        <div key={i} style={{ padding: 8, ...(i < items.length - 1 ? { borderBottom: '8px solid var(--bg-dark)' } : {}) }}>
          {child}
        </div>
      ))}
    </div>
  );
};

export const TextContent: React.FC<React.PropsWithChildren> = ({ children }) => (
  <div style={{ padding: 10, fontSize: 60 }}>{children}</div>
);

export const AdminFooter: React.FC<React.PropsWithChildren> = ({ children }) => (
  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', backgroundColor: 'var(--bg-dark)', color: 'var(--text)', height: 100, padding: '0 32px', width: '100%' }}>{children}</div>
);

export const FooterItem: React.FC<React.PropsWithChildren<{ className?: string; style?: React.CSSProperties }>> = ({
  className,
  children,
  style,
}) => (
  <div className={className} style={{ display: 'flex', flexDirection: 'row', fontSize: 40, gap: 32, alignItems: 'center', ...style }}>{children}</div>
);

export const GameAdmin: React.FC<React.PropsWithChildren> = ({ children }) => (
  <div style={{ height: '100%', display: 'flex', flexDirection: 'column', backgroundColor: 'var(--bg-base)' }}>
    {children}
    <Toast />
  </div>
);
