import { useTheme } from '@/contexts/ThemeContext';
import { motion, useReducedMotion } from 'motion/react';

interface AuthCardProps {
  children: React.ReactNode;
  title: string;
}

const AuthCard = ({ children, title }: AuthCardProps) => {
  const { darkMode } = useTheme();
  const reduceMotion = useReducedMotion();

  return (
    <main
      className={[
        darkMode ? 'dark' : '',
        'grid min-h-screen place-items-center bg-linkflow-page px-4 py-6 text-linkflow-text transition-colors duration-300 [perspective:1200px] sm:px-6 lg:px-8',
        'bg-[linear-gradient(180deg,var(--color-linkflow-page-tint),rgba(255,255,255,0)_280px)]',
        'dark:bg-linkflow-dark-page dark:text-linkflow-dark-text',
        'dark:bg-[linear-gradient(180deg,var(--color-linkflow-dark-page-tint),rgba(11,16,32,0)_280px)]',
      ].join(' ')}
    >
      <motion.div
        initial={reduceMotion ? false : { rotateY: -180 }}
        animate={reduceMotion ? undefined : { rotateY: 0 }}
        transition={{ duration: 0.3, ease: 'easeOut' }}
        className="w-full max-w-md [transform-origin:center] [transform-style:preserve-3d]"
      >
        <motion.section
          initial={reduceMotion ? false : { opacity: 0 }}
          animate={reduceMotion ? undefined : { opacity: [0, 0, 1] }}
          transition={
            reduceMotion
              ? undefined
              : { duration: 0.3, ease: 'easeOut', times: [0, 0.48, 1] }
          }
          className={[
            'w-full rounded-lg border border-linkflow-border bg-linkflow-panel p-6 shadow-xl shadow-slate-200/70',
            'transition duration-300 [backface-visibility:hidden] [transform-style:preserve-3d] hover:border-linkflow-primary hover:shadow-2xl hover:shadow-linkflow-primary-soft focus-within:border-linkflow-primary focus-within:shadow-2xl focus-within:shadow-linkflow-primary-soft sm:p-7',
            'dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/40 dark:hover:border-linkflow-dark-primary dark:hover:shadow-linkflow-dark-primary-soft dark:focus-within:border-linkflow-dark-primary dark:focus-within:shadow-linkflow-dark-primary-soft',
          ].join(' ')}
          aria-labelledby="auth-title"
        >
          <div className="mb-6 text-center">
            <div className="mb-3 flex items-center justify-center gap-2">
              <span className="grid h-8 w-8 place-items-center rounded-md bg-linkflow-primary text-sm font-bold text-white shadow-md shadow-linkflow-primary-soft">
                LF
              </span>
              <span className="text-lg font-bold tracking-normal text-linkflow-text dark:text-linkflow-dark-text">
                LinkFlow
              </span>
            </div>
            <h1 id="auth-title" className="text-2xl font-bold leading-tight">
              {title}
            </h1>
          </div>

          {children}
        </motion.section>
      </motion.div>
    </main>
  );
};

export default AuthCard;
