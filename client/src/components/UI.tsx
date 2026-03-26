import React from 'react';
import { ToastContainer } from 'react-toastify';
import { Link } from 'react-router-dom';
import 'react-toastify/dist/ReactToastify.css';

export const Loading: React.FC = () => <div className="loading">Loading...</div>;

export const Toast: React.FC = () => (
  <ToastContainer
    position="top-right"
    autoClose={3000}
    hideProgressBar
    newestOnTop={false}
    closeOnClick
    rtl={false}
    pauseOnFocusLoss
    draggable
    pauseOnHover
  />
);

type BtnProps = React.PropsWithChildren<{
  onClick?: () => void;
  className?: string;
  style?: React.CSSProperties;
}>;

export const Button: React.FC<BtnProps> = ({ onClick, className = '', children, style }) => (
  <div className={`btn ${className}`} onClick={onClick} style={style}>{children}</div>
);

export const OvalButton: React.FC<BtnProps> = ({ onClick, className = '', children }) => (
  <div className={`btn btn-oval ${className}`} onClick={onClick}>{children}</div>
);

export const ButtonLink: React.FC<React.PropsWithChildren<{ to: string; className?: string }>> = ({
  to,
  className = '',
  children,
}) => (
  <Link className={`btn ${className}`} to={to}>{children}</Link>
);

export const Input: React.FC<{
  type?: string;
  onChange?: React.ChangeEventHandler<HTMLInputElement>;
  value?: string | number;
  className?: string;
  style?: React.CSSProperties;
}> = ({ type, onChange, value, className = '', style }) => (
  <input className={`input ${className}`} type={type} onChange={onChange} value={value} style={style} />
);

export const VerticalList: React.FC<React.PropsWithChildren<{ className?: string; style?: React.CSSProperties }>> = ({
  className = '',
  style,
  children,
}) => <div className={`list list-vertical ${className}`} style={style}>{children}</div>;

export const HorizontalList: React.FC<React.PropsWithChildren<{ className?: string; style?: React.CSSProperties }>> = ({
  className = '',
  style,
  children,
}) => <div className={`list list-horizontal ${className}`} style={style}>{children}</div>;

export const ListItem: React.FC<
  React.PropsWithChildren<{ className?: string; style?: React.CSSProperties; onClick?: () => void }>
> = ({ className = '', style, children, ...props }) => (
  <div className={`list-item ${className}`} style={style} {...props}>{children}</div>
);

export const TwoLineListItem: React.FC<
  React.PropsWithChildren<{ className?: string; style?: React.CSSProperties; onClick?: () => void }>
> = ({ className = '', style, children, ...props }) => (
  <div className={`list-item list-item-two-line ${className}`} style={style} {...props}>{children}</div>
);
