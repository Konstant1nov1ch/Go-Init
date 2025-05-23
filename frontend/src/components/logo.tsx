import { Box } from '@mui/material';
import logoSvg from '../assets/logo/logog.svg';
import logoPng from '../assets/logo/logog.png';

interface LogoProps {
  size?: number;
  png?: boolean;        // true → PNG-версия (герой в сайдбаре)
  className?: string;
}

export const Logo = ({ size = 48, png = false, className }: LogoProps) => (
  <Box
    component="img"
    src={png ? logoPng : logoSvg}
    alt="Go init logo"
    className={className}
    sx={{ width: size, height: size }}
  />
);
