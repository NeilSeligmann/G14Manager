// The following ifdef block is the standard way of creating macros which make exporting
// from a DLL simpler. All files within this DLL are compiled with the MATRIXCONTROLLER_EXPORTS
// symbol defined on the command line. This symbol should not be defined on any project
// that uses this DLL. This way any other project whose source files include this file see
// MATRIXCONTROLLER_API functions as being imported from a DLL, whereas this DLL sees symbols
// defined with this macro as being exported.
#ifdef MATRIXCONTROLLER_EXPORTS
#define MATRIXCONTROLLER_API __declspec(dllexport)
#else
#define MATRIXCONTROLLER_API __declspec(dllimport)
#endif

#include "MatrixController.h"

typedef struct _API_WRAPPER {
	WINUSB_INTERFACE_HANDLE WinusbHandle;
	HANDLE                  DeviceHandle;

	MatrixController*		mc;
	char					devicePath[MAX_PATH + 1];
} API_WRAPPER, *PAPI_WRAPPER;

#ifdef __cplusplus
extern "C" {
#endif

	MATRIXCONTROLLER_API PAPI_WRAPPER NewController(void);
	MATRIXCONTROLLER_API void DeleteController(PAPI_WRAPPER w);

	MATRIXCONTROLLER_API int PrepareDraw(PAPI_WRAPPER w, unsigned char* m, size_t len);
	MATRIXCONTROLLER_API int DrawMatrix(PAPI_WRAPPER w);
	MATRIXCONTROLLER_API int ClearMatrix(PAPI_WRAPPER w);

#ifdef __cplusplus
}
#endif
