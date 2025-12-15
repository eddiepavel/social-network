import { useState } from 'react';
import type { ChangeEvent, FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { ErrorMessage } from '../common/ErrorMessage';
import type { RegisterFormData } from '../../types';

export function RegisterForm() {
  const navigate = useNavigate();
  const { register, isLoading, clearError } = useAuth();

  const [formData, setFormData] = useState<RegisterFormData>({
    email: '',
    password: '',
    first_name: '',
    last_name: '',
    dob: '',
    nickname: '',
    about_me: '',
  });
  const [localError, setLocalError] = useState<string>('');

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLocalError('');
    clearError();

    // Validation
    if (!formData.email || !formData.password || !formData.first_name || !formData.last_name || !formData.dob) {
      setLocalError('Please fill in all required fields');
      return;
    }

    if (!formData.email.includes('@')) {
      setLocalError('Please enter a valid email address');
      return;
    }

    if (formData.password.length < 6) {
      setLocalError('Password must be at least 6 characters long');
      return;
    }

    // Validate date of birth (user must be at least 13 years old)
    const today = new Date();
    const birthDate = new Date(formData.dob);
    const age = today.getFullYear() - birthDate.getFullYear();
    const monthDiff = today.getMonth() - birthDate.getMonth();

    if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birthDate.getDate())) {
      // Adjust age if birthday hasn't occurred this year
    }

    if (age < 13) {
      setLocalError('You must be at least 13 years old to register');
      return;
    }

    try {
      // Remove empty optional fields
      const cleanedData = {
        ...formData,
        nickname: formData.nickname || undefined,
        about_me: formData.about_me || undefined,
      };

      await register(cleanedData);
      // On success, navigate to dashboard
      navigate('/dashboard');
    } catch (error) {
      if (error instanceof Error) {
        setLocalError(error.message);
      } else {
        setLocalError('Registration failed. Please try again.');
      }
    }
  };

  const handleChange = (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
    // Clear errors when user starts typing
    if (localError) setLocalError('');
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="bg-card/80 border border-white/10 rounded-2xl p-8 shadow-soft backdrop-blur"
    >
      <div className="flex items-center justify-between mb-6">
        <div>
          <p className="text-sm text-muted uppercase tracking-wide">Create account</p>
          <h2 className="text-2xl font-semibold mt-1">Register</h2>
        </div>
      </div>

      {localError && <ErrorMessage message={localError} onDismiss={() => setLocalError('')} />}

      <div className="space-y-6">
        <div className="space-y-2">
          <label htmlFor="email" className="text-sm text-muted">
            Email *
          </label>
        <input
          type="email"
          id="email"
          name="email"
          value={formData.email}
          onChange={handleChange}
          required
          disabled={isLoading}
          placeholder="Enter your email"
        />
        </div>

        <div className="space-y-2">
          <label htmlFor="password" className="text-sm text-muted">
            Password *
          </label>
        <input
          type="password"
          id="password"
          name="password"
          value={formData.password}
          onChange={handleChange}
          required
          disabled={isLoading}
          placeholder="At least 6 characters"
          minLength={6}
        />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <label htmlFor="first_name" className="text-sm text-muted">
              First Name *
            </label>
          <input
            type="text"
            id="first_name"
            name="first_name"
            value={formData.first_name}
            onChange={handleChange}
            required
            disabled={isLoading}
            placeholder="First name"
          />
          </div>

          <div className="space-y-2">
            <label htmlFor="last_name" className="text-sm text-muted">
              Last Name *
            </label>
          <input
            type="text"
            id="last_name"
            name="last_name"
            value={formData.last_name}
            onChange={handleChange}
            required
            disabled={isLoading}
            placeholder="Last name"
          />
          </div>
        </div>

        <div className="space-y-2">
          <label htmlFor="dob" className="text-sm text-muted">
            Date of Birth *
          </label>
        <input
          type="date"
          id="dob"
          name="dob"
          value={formData.dob}
          onChange={handleChange}
          required
          disabled={isLoading}
        />
        </div>

        <div className="space-y-2">
          <label htmlFor="nickname" className="text-sm text-muted">
            Nickname (Optional)
          </label>
        <input
          type="text"
          id="nickname"
          name="nickname"
          value={formData.nickname}
          onChange={handleChange}
          disabled={isLoading}
          placeholder="Your nickname"
        />
        </div>

        <div className="space-y-2">
          <label htmlFor="about_me" className="text-sm text-muted">
            About Me (Optional)
          </label>
        <textarea
          id="about_me"
          name="about_me"
          value={formData.about_me}
          onChange={handleChange}
          disabled={isLoading}
          placeholder="Tell us about yourself"
          rows={3}
        />
        </div>
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="mt-8 w-full rounded-xl bg-primary-600 hover:bg-primary-500 transition text-white font-semibold py-3 disabled:opacity-60 disabled:cursor-not-allowed shadow-lg shadow-primary-900/40"
      >
        {isLoading ? 'Creating Account...' : 'Register'}
      </button>

      <p className="mt-4 text-center text-sm text-muted">
        Already have an account?{' '}
        <a href="/login" className="font-semibold text-primary-300 hover:text-primary-200">
          Login
        </a>
      </p>
    </form>
  );
}
