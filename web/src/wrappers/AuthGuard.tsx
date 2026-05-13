import { useAuthStore } from '@/stores/authStore';
import { Navigate, Outlet, useLocation } from '@umijs/max';

const AuthGuard = () => {
  const accessToken = useAuthStore((state) => state.accessToken);
  const location = useLocation();

  if (!accessToken) {
    return (
      <Navigate
        to={`/login?redirect=${encodeURIComponent(
          location.pathname + location.search,
        )}`}
        replace
      />
    );
  }

  return <Outlet />;
};

export default AuthGuard;
