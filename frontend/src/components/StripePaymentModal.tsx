"use client";

import { useState } from "react";
import { loadStripe } from "@stripe/stripe-js";
import {
  Elements,
  PaymentElement,
  useStripe,
  useElements,
} from "@stripe/react-stripe-js";
import { X, Loader2, CreditCard, CheckCircle2 } from "lucide-react";

const stripePublishableKey =
  process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY ||
  "pk_test_TYooMQauvdEDq54NiTphI7jx";

const stripePromise = loadStripe(stripePublishableKey);

interface FormProps {
  appointmentTitle: string;
  amount: number;
  currency: string;
  onSuccess: () => void;
  onClose: () => void;
}

function CheckoutForm({ appointmentTitle, amount, currency, onSuccess, onClose }: FormProps) {
  const stripe = useStripe();
  const elements = useElements();
  const [isProcessing, setIsProcessing] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const [isPaidSuccess, setIsPaidSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!stripe || !elements) {
      return;
    }

    setIsProcessing(true);
    setErrorMessage("");

    const { error, paymentIntent } = await stripe.confirmPayment({
      elements,
      confirmParams: {
        return_url: window.location.href,
      },
      redirect: "if_required",
    });

    if (error) {
      setErrorMessage(error.message || "An unexpected error occurred.");
      setIsProcessing(false);
    } else if (paymentIntent && paymentIntent.status === "succeeded") {
      setIsPaidSuccess(true);
      setIsProcessing(false);
      setTimeout(() => {
        onSuccess();
        onClose();
      }, 2000);
    } else {
      setIsProcessing(false);
    }
  };

  if (isPaidSuccess) {
    return (
      <div className="py-8 text-center flex flex-col items-center">
        <CheckCircle2 size={48} className="text-emerald-400 mb-3 animate-bounce" />
        <h3 className="text-xl font-bold text-white mb-1">Payment Successful!</h3>
        <p className="text-sm text-gray-400">Your appointment has been successfully paid.</p>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="bg-white/5 p-4 rounded-xl border border-white/10 mb-4">
        <p className="text-xs text-gray-400 font-medium uppercase tracking-wider mb-1">Item Details</p>
        <div className="flex justify-between items-center">
          <span className="font-semibold text-white">{appointmentTitle}</span>
          <span className="text-lg font-extrabold text-blue-400">
            {(amount / 100).toLocaleString("en-US", { style: "currency", currency: currency.toUpperCase() })}
          </span>
        </div>
      </div>

      {errorMessage && (
        <div className="p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm">
          {errorMessage}
        </div>
      )}

      <div className="p-4 bg-black/40 rounded-xl border border-white/10 [color-scheme:dark]">
        <PaymentElement />
      </div>

      <div className="flex justify-end gap-3 pt-2">
        <button
          type="button"
          onClick={onClose}
          disabled={isProcessing}
          className="px-4 py-2.5 text-sm font-medium text-gray-400 hover:text-white transition-colors"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={!stripe || isProcessing}
          className="px-6 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-semibold rounded-lg transition-colors flex items-center gap-2 shadow-[0_0_20px_rgba(16,185,129,0.3)] disabled:opacity-70"
        >
          {isProcessing ? (
            <>
              <Loader2 className="animate-spin" size={18} /> Processing...
            </>
          ) : (
            `Pay ${(amount / 100).toLocaleString("en-US", { style: "currency", currency: currency.toUpperCase() })}`
          )}
        </button>
      </div>
    </form>
  );
}

interface ModalProps {
  isOpen: boolean;
  clientSecret: string;
  appointmentTitle: string;
  amount: number;
  currency: string;
  onSuccess: () => void;
  onClose: () => void;
}

export default function StripePaymentModal({
  isOpen,
  clientSecret,
  appointmentTitle,
  amount,
  currency,
  onSuccess,
  onClose,
}: ModalProps) {
  if (!isOpen || !clientSecret) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-md p-4">
      <div className="glass-panel w-full max-w-lg p-6 relative bg-black/90">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-gray-400 hover:text-white transition-colors"
        >
          <X size={20} />
        </button>

        <div className="flex items-center gap-3 mb-6">
          <div className="p-3 bg-emerald-500/10 text-emerald-400 rounded-lg">
            <CreditCard size={24} />
          </div>
          <div>
            <h2 className="text-xl font-bold">Secure Checkout</h2>
            <p className="text-sm text-gray-400">Complete your payment via Stripe</p>
          </div>
        </div>

        <Elements stripe={stripePromise} options={{ clientSecret, appearance: { theme: 'night' } }}>
          <CheckoutForm
            appointmentTitle={appointmentTitle}
            amount={amount}
            currency={currency}
            onSuccess={onSuccess}
            onClose={onClose}
          />
        </Elements>
      </div>
    </div>
  );
}
