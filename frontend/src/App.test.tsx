import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import App from './App';
import * as authApi from './api/auth';
import * as healthClient from './api/client';

describe('App & Authentication Flow', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    localStorage.clear();

    // Default health response mock
    vi.spyOn(healthClient, 'fetchHealth').mockResolvedValue({
      success: true,
      status: 'healthy',
      database: 'connected',
    });
  });

  it('renders login form when unauthenticated', async () => {
    render(<App />);

    await waitFor(() => {
      expect(screen.getByText(/Welcome back/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Sign In/i })).toBeInTheDocument();
    });
  });

  it('shows client-side validation errors for invalid input', async () => {
    render(<App />);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Sign In/i })).toBeInTheDocument();
    });

    const submitBtn = screen.getByRole('button', { name: /Sign In/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByText(/Please enter a valid email address/i)).toBeInTheDocument();
    });

    const emailInput = screen.getByLabelText(/Email Address/i);
    fireEvent.change(emailInput, { target: { value: 'user@example.com' } });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByText(/Password must be at least 8 characters long/i)).toBeInTheDocument();
    });
  });

  it('toggles to register mode and displays cycle start day field', async () => {
    render(<App />);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Create an account/i })).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /Create an account/i }));

    await waitFor(() => {
      expect(screen.getByText(/Create your account/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/Financial Cycle Start Day/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Create Account/i })).toBeInTheDocument();
    });
  });

  it('logs in successfully and renders authenticated dashboard', async () => {
    vi.spyOn(authApi, 'login').mockResolvedValue({
      token: 'valid-test-token',
      profile: {
        id: '12345678-1234-1234-1234-123456789abc',
        email: 'investor@example.com',
        cycle_start_day: 25,
        created_at: '2026-10-04T00:00:00Z',
      },
    });

    render(<App />);

    await waitFor(() => {
      expect(screen.getByLabelText(/Email Address/i)).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText(/Email Address/i), {
      target: { value: 'investor@example.com' },
    });
    fireEvent.change(screen.getByLabelText(/Password/i), {
      target: { value: 'Secret123!' },
    });

    fireEvent.click(screen.getByRole('button', { name: /Sign In/i }));

    await waitFor(() => {
      expect(screen.getByText(/Welcome, investor@example.com!/i)).toBeInTheDocument();
      expect(screen.getByText(/Day 25 of each month/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Sign Out/i })).toBeInTheDocument();
    });

    // Test Sign Out
    vi.spyOn(authApi, 'logout').mockResolvedValue();
    fireEvent.click(screen.getByRole('button', { name: /Sign Out/i }));

    await waitFor(() => {
      expect(screen.getByText(/Welcome back/i)).toBeInTheDocument();
    });
  });

  it('updates cycle start day in settings card', async () => {
    vi.spyOn(authApi, 'getToken').mockReturnValue('stored-token');
    vi.spyOn(authApi, 'getMe').mockResolvedValue({
      id: '12345678-1234-1234-1234-123456789abc',
      email: 'investor@example.com',
      cycle_start_day: 1,
      created_at: '2026-10-04T00:00:00Z',
    });

    const updateSpy = vi.spyOn(authApi, 'updateUserSettings').mockResolvedValue({
      cycle_start_day: 15,
    });

    render(<App />);

    await waitFor(() => {
      expect(screen.getByText(/Welcome, investor@example.com!/i)).toBeInTheDocument();
    });

    // Navigate to profile & settings
    fireEvent.click(screen.getByRole('button', { name: /Profile & Settings/i }));

    const cycleInput = screen.getByLabelText(/Cycle Start Day/i);
    fireEvent.change(cycleInput, { target: { value: '15' } });

    const saveBtn = screen.getByRole('button', { name: /Save Changes/i });
    expect(saveBtn).not.toBeDisabled();
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(updateSpy).toHaveBeenCalledWith(15);
      expect(screen.getByText(/Financial cycle start day updated successfully!/i)).toBeInTheDocument();
    });
  });
});
