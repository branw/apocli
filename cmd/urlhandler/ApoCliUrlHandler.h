#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>

extern void HandleURL(char*);

@interface BrowseAppDelegate: NSObject<NSApplicationDelegate>
  - (void)handleGetURLEvent:(NSAppleEventDescriptor *) event withReplyEvent:(NSAppleEventDescriptor *)replyEvent;
@end

void RunApp();
int ShowAlert(bool, char*, char*);
