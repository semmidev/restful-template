import useAuthStore from '../features/auth/store';

export function usePermission(permission: string): boolean {
  const permissions = useAuthStore((state) => state.permissions) || [];
  return permissions.includes(permission);
}

export default usePermission;
