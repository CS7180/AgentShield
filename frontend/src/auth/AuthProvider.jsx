import React from 'react';
import { useEffect, useState } from 'react';
import AuthContext from './AuthContext';
import { isSupabaseConfigured, supabase } from '../lib/supabase';

export function AuthProvider({ children }) {
  const [session, setSession] = useState(null);
  const [loading, setLoading] = useState(isSupabaseConfigured);
  const [authError, setAuthError] = useState('');

  useEffect(() => {
    if (!isSupabaseConfigured || !supabase) {
      setLoading(false);
      setAuthError('');
      return undefined;
    }

    let isMounted = true;

    supabase.auth
      .getSession()
      .then(({ data, error }) => {
        if (!isMounted) return;

        if (error) {
          console.error('Failed to load Supabase session', error);
          setSession(null);
          setAuthError(error.message || 'Failed to restore Supabase session.');
          setLoading(false);
          return;
        }

        setSession(data.session ?? null);
        setAuthError('');
        setLoading(false);
      })
      .catch((error) => {
        if (!isMounted) return;
        console.error('Unexpected Supabase session bootstrap failure', error);
        setSession(null);
        setAuthError(error?.message || 'Unexpected auth initialization failure.');
        setLoading(false);
      });

    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange((_event, nextSession) => {
      setSession(nextSession);
      setAuthError('');
      setLoading(false);
    });

    return () => {
      isMounted = false;
      subscription.unsubscribe();
    };
  }, []);

  const value = {
    session,
    user: session?.user ?? null,
    loading,
    isConfigured: isSupabaseConfigured,
    authError,
    async signInWithGoogle({ redirectTo }) {
      if (!supabase) {
        return { error: new Error('Supabase is not configured yet.') };
      }

      return supabase.auth.signInWithOAuth({
        provider: 'google',
        options: {
          redirectTo,
        },
      });
    },
    async signOut() {
      if (!supabase) {
        return { error: null };
      }

      return supabase.auth.signOut();
    },
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
