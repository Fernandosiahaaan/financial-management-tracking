import React, { createContext, useContext, useState, type ReactNode } from 'react';
import { formatRupiah, formatNumberIDR } from '../utils/currency';

export const MASKED_AMOUNT = 'Rp ••••••••';
export const MASKED_NUMBER = '••••••••';

export interface PrivacyContextType {
  isHidden: boolean;
  togglePrivacy: () => void;
  setIsHidden: (hidden: boolean) => void;
  formatAmount: (amountInCents: number) => string;
  formatNumber: (amountInCents: number) => string;
  maskValue: (formattedValue: string) => string;
}

const defaultContextValue: PrivacyContextType = {
  isHidden: false, // Fallback for standalone test renders without provider
  togglePrivacy: () => {},
  setIsHidden: () => {},
  formatAmount: (amountInCents: number) => formatRupiah(amountInCents),
  formatNumber: (amountInCents: number) => formatNumberIDR(amountInCents),
  maskValue: (formattedValue: string) => formattedValue,
};

const PrivacyContext = createContext<PrivacyContextType>(defaultContextValue);

export interface PrivacyProviderProps {
  children: ReactNode;
  defaultHidden?: boolean;
}

export const PrivacyProvider: React.FC<PrivacyProviderProps> = ({
  children,
  defaultHidden = true,
}) => {
  const [isHidden, setIsHidden] = useState<boolean>(defaultHidden);

  const togglePrivacy = () => {
    setIsHidden((prev) => !prev);
  };

  const formatAmount = (amountInCents: number): string => {
    if (isHidden) {
      return amountInCents < 0 ? `-${MASKED_AMOUNT}` : MASKED_AMOUNT;
    }
    return formatRupiah(amountInCents);
  };

  const formatNumber = (amountInCents: number): string => {
    if (isHidden) {
      return amountInCents < 0 ? `-${MASKED_NUMBER}` : MASKED_NUMBER;
    }
    return formatNumberIDR(amountInCents);
  };

  const maskValue = (formattedValue: string): string => {
    if (!isHidden) return formattedValue;
    if (formattedValue.startsWith('-Rp ') || formattedValue.startsWith('- Rp ')) {
      return '-Rp ••••••••';
    }
    if (formattedValue.startsWith('+Rp ') || formattedValue.startsWith('+ Rp ')) {
      return '+Rp ••••••••';
    }
    if (formattedValue.startsWith('Rp ') || formattedValue.startsWith('Rp')) {
      return 'Rp ••••••••';
    }
    if (formattedValue.startsWith('-')) {
      return '-••••••••';
    }
    if (formattedValue.startsWith('+')) {
      return '+••••••••';
    }
    return '••••••••';
  };

  return (
    <PrivacyContext.Provider
      value={{
        isHidden,
        togglePrivacy,
        setIsHidden,
        formatAmount,
        formatNumber,
        maskValue,
      }}
    >
      {children}
    </PrivacyContext.Provider>
  );
};

export const usePrivacy = (): PrivacyContextType => {
  return useContext(PrivacyContext);
};
