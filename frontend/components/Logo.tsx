import React from 'react';

interface LogoProps {
  size?: 'small' | 'medium' | 'large';
  withText?: boolean;
  className?: string;
}

const Logo: React.FC<LogoProps> = ({ 
  size = 'medium', 
  withText = true, 
  className = '' 
}) => {
  const sizeClasses = {
    small: 'logo--small',
    medium: 'logo--medium',
    large: 'logo--large'
  };

  return (
    <img 
      src={withText ? '/images/logo_full.png' : '/favicon-32x32.png'}
      alt="Social Network Logo"
      className={`logo ${sizeClasses[size]} ${className}`}
    />
  );
};

export default Logo;