import type { ReactNode } from 'react';
import styles from './styles.module.css';
import clsx from 'clsx';

interface LinkButtonProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  children: ReactNode;
  className?: string;
}

export function LinkButton({ children, className, ...props }: LinkButtonProps): ReactNode {
  return (
    <a
      className={clsx(styles.linkButton, className)}
      {...props}
    >
      {children}
    </a>
  );
}
