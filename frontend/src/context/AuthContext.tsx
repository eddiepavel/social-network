import { createContext, useReducer } from 'react';
import type { ReactNode } from 'react';
import type { AuthState, LoginFormData, RegisterFormData, User } from '../types';
import { authReducer, initialAuthState } from './authReducer';
import * as authService from '../services/authService';

// Context type definition
interface AuthContextType extends AuthState {
  login: (credentials: LoginFormData) => Promise<void>;
  register: (formData: RegisterFormData) => Promise<void>;
  logout: () => Promise<void>;
  checkSession: () => Promise<void>;
  clearError: () => void;
  setUser: (user: User) => void;
}

// Create context with undefined default (will be provided by AuthProvider)
export const AuthContext = createContext<AuthContextType | undefined>(undefined);

// AuthProvider props
interface AuthProviderProps {
  children: ReactNode;
}

// AuthProvider component
export function AuthProvider({ children }: AuthProviderProps) {
  const [state, dispatch] = useReducer(authReducer, initialAuthState);

  // Login function
  const login = async (credentials: LoginFormData) => {
    dispatch({ type: 'LOGIN_START' });
    try {
      const user: User = await authService.login(credentials);
      dispatch({ type: 'LOGIN_SUCCESS', payload: user });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Login failed';
      dispatch({ type: 'LOGIN_FAILURE', payload: errorMessage });
      throw error; // Re-throw so component can handle it
    }
  };

  // Register function
  const register = async (formData: RegisterFormData) => {
    dispatch({ type: 'REGISTER_START' });
    try {
      const user: User = await authService.register(formData);
      dispatch({ type: 'REGISTER_SUCCESS', payload: user });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Registration failed';
      dispatch({ type: 'REGISTER_FAILURE', payload: errorMessage });
      throw error; // Re-throw so component can handle it
    }
  };

  // Logout function
  const logout = async () => {
    try {
      await authService.logout();
      dispatch({ type: 'LOGOUT' });
    } catch (error) {
      // Even if API fails, clear local state
      dispatch({ type: 'LOGOUT' });
      console.error('Logout error:', error);
    }
  };

  // Check session function (for auto-login)
  const checkSession = async () => {
    try {
      const user: User = await authService.getSession();
      dispatch({ type: 'SET_USER', payload: user });
    } catch (error) {
      // Session invalid or expired, clear state
      dispatch({ type: 'LOGOUT' });
    }
  };

  // Clear error function
  const clearError = () => {
    dispatch({ type: 'CLEAR_ERROR' });
  };

  // Set user function (for updating user data)
  const setUser = (user: User) => {
    dispatch({ type: 'SET_USER', payload: user });
  };

  const value: AuthContextType = {
    ...state,
    login,
    register,
    logout,
    checkSession,
    clearError,
    setUser,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
