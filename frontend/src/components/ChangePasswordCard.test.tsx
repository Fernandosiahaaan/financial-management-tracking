import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ChangePasswordCard } from './ChangePasswordCard';
import * as authApi from '../api/auth';

describe('ChangePasswordCard Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders all form fields and submit button', () => {
    render(<ChangePasswordCard />);

    expect(screen.getByText('Ganti Kata Sandi')).toBeInTheDocument();
    expect(screen.getByLabelText(/Kata Sandi Saat Ini/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^Kata Sandi Baru$/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Konfirmasi Kata Sandi Baru/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Perbarui Kata Sandi/i })).toBeDisabled();
  });

  it('validates password mismatch on client side', async () => {
    render(<ChangePasswordCard />);

    const currentInput = screen.getByLabelText(/Kata Sandi Saat Ini/i);
    const newInput = screen.getByLabelText(/^Kata Sandi Baru$/i);
    const confirmInput = screen.getByLabelText(/Konfirmasi Kata Sandi Baru/i);
    const submitBtn = screen.getByRole('button', { name: /Perbarui Kata Sandi/i });

    fireEvent.change(currentInput, { target: { value: 'OldPassword123!' } });
    fireEvent.change(newInput, { target: { value: 'NewPassword123!' } });
    fireEvent.change(confirmInput, { target: { value: 'Mismatch123!' } });

    expect(submitBtn).toBeEnabled();
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByText('Konfirmasi kata sandi baru tidak cocok')).toBeInTheDocument();
    });
  });

  it('validates new password min length on client side', async () => {
    render(<ChangePasswordCard />);

    const currentInput = screen.getByLabelText(/Kata Sandi Saat Ini/i);
    const newInput = screen.getByLabelText(/^Kata Sandi Baru$/i);
    const confirmInput = screen.getByLabelText(/Konfirmasi Kata Sandi Baru/i);
    const submitBtn = screen.getByRole('button', { name: /Perbarui Kata Sandi/i });

    fireEvent.change(currentInput, { target: { value: 'OldPassword123!' } });
    fireEvent.change(newInput, { target: { value: 'short' } });
    fireEvent.change(confirmInput, { target: { value: 'short' } });

    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByText('Kata sandi baru minimal 8 karakter')).toBeInTheDocument();
    });
  });

  it('submits change password successfully and resets form', async () => {
    const changeSpy = vi.spyOn(authApi, 'changePassword').mockResolvedValue();
    render(<ChangePasswordCard />);

    const currentInput = screen.getByLabelText(/Kata Sandi Saat Ini/i) as HTMLInputElement;
    const newInput = screen.getByLabelText(/^Kata Sandi Baru$/i) as HTMLInputElement;
    const confirmInput = screen.getByLabelText(/Konfirmasi Kata Sandi Baru/i) as HTMLInputElement;
    const submitBtn = screen.getByRole('button', { name: /Perbarui Kata Sandi/i });

    fireEvent.change(currentInput, { target: { value: 'OldPassword123!' } });
    fireEvent.change(newInput, { target: { value: 'NewPassword123!' } });
    fireEvent.change(confirmInput, { target: { value: 'NewPassword123!' } });

    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(changeSpy).toHaveBeenCalledWith('OldPassword123!', 'NewPassword123!');
      expect(screen.getByText('Kata sandi berhasil diperbarui!')).toBeInTheDocument();
      expect(currentInput.value).toBe('');
      expect(newInput.value).toBe('');
      expect(confirmInput.value).toBe('');
    });
  });
});
