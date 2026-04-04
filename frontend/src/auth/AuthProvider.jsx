import { useEffect, useState } from 'react';
import AuthContext from './AuthContext';
import { isSupabaseConfigured, supabase } from '../lib/supabase';

export function AuthProvider({ children }) {
  const [session, setSession] = useState(null);
  const [loading, setLoading] = useState(isSupabaseConfigured);

  useEffect(() => {
    if (!isSupabaseConfigured || !supabase) {
      return undefined;
    }

    let isMounted = true;

    supabase.auth.getSession().then(({ data, error }) => {
      if (error) {
        console.error('Failed to load Supabase session', error);
      }

      if (isMounted) {
        setSession(data.session ?? null);
        setLoading(false);
      }
    });

    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange((_event, nextSession) => {
      setSession(nextSession);
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
