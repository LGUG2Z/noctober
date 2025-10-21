import React, { useState, useEffect } from "react";
import Navbar from "../components/Navbar";
import { toast } from "react-hot-toast";
import { BrowserOpenURL } from "../../wailsjs/runtime";
import {
  GetSettings,
  NavigateExplorerToLogLocation,
  FormatSystemDetails,
  GetPlainSystemDetails,
} from "../../wailsjs/go/backend/Backend";
import {
  SaveToken,
} from "../../wailsjs/go/backend/Settings";

export default function Settings() {
  const [loaded, setLoadState] = useState(false);
  const [token, setToken] = useState("");
  const [storeHighlights, setStoreHighlights] = useState(false);
  const [tokenInput, setTokenInput] = useState("");
  const [systemDetails, setSystemDetails] = useState(
    "Fetching system details...",
  );

  useEffect(() => {
    GetSettings().then((settings) => {
      setLoadState(true);
      setToken(settings.notado_token);
      setTokenInput(settings.notado_token);
      setStoreHighlights(settings.upload_store_highlights);
    });
    GetPlainSystemDetails().then((details) => setSystemDetails(details));
  }, [loaded]);

  function saveToken() {
    setToken(tokenInput);
    SaveToken(tokenInput);
    toast.success("Your changes have been saved");
  }

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-800 flex flex-col">
      <Navbar />
      <div className="flex-grow items-center justify-center pb-24 px-24 grid grid-cols-2 gap-14">
        <div className="space-y-2">
          <h2 className="text-center text-3xl font-extrabold text-gray-900 dark:text-gray-300">
            Settings
          </h2>
        </div>
        <div className="space-y-4">
          <div className="bg-white dark:bg-slate-700 shadow sm:rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-gray-300">
                Set your Notado access token
              </h3>
              <div className="mt-2 max-w-xl text-sm text-gray-500 dark:text-gray-400">
                <p>
                  You can find your access token at{" "}
                  <button
                    className="text-gray-600 dark:text-gray-400 underline"
                    onClick={() =>
                      BrowserOpenURL("https://notado.app/settings")
                    }
                  >
                    https://notado.app/settings
                  </button>
                </p>
              </div>
              <form
                onSubmit={(e) => e.preventDefault()}
                className="sm:flex flex-col"
              >
                <div className="w-full mt-4 sm:flex sm:items-center">
                  <input
                    onChange={(e) => setTokenInput(e.target.value)}
                    type="text"
                    name="token"
                    id="token"
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 dark:bg-gray-200 focus:bg-white rounded-md"
                    placeholder="Your access token goes here"
                    value={tokenInput}
                  />
                </div>
                <div className="w-full mt-4 sm:flex flex-row">
                  <button
                    onClick={saveToken}
                    type="submit"
                    className="mt-3 w-full inline-flex items-center justify-center px-4 py-2 border border-transparent shadow-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:text-sm"
                  >
                    Save
                  </button>
                </div>
              </form>
            </div>
          </div>
          <div className="shadow overflow-hidden sm:rounded-md">
            <div className="px-4 py-5 bg-white dark:bg-slate-700 space-y-6 sm:p-6">
              <fieldset>
                <legend className="text-base font-medium text-gray-900 dark:text-gray-300">
                  Having trouble?
                </legend>
                <div className="mt-2 max-w-xl text-sm text-gray-500 dark:text-gray-400">
                  <p>Noctober Build: {systemDetails}</p>
                </div>
                <div className="space-y-4">
                  <div className="flex items-start">
                    <div className="w-full mt-4 sm:flex flex-row">
                      <button
                        onClick={NavigateExplorerToLogLocation}
                        type="submit"
                        className="mt-3 w-full inline-flex items-center justify-center px-4 py-2 border border-transparent shadow-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:text-sm"
                      >
                        Open Logs Folder
                      </button>
                    </div>
                  </div>
                </div>
              </fieldset>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
