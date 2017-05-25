#!/usr/bin/perl

package TSCollector;

require Exporter;

use LWP::UserAgent;
use HTTP::Request::Common;
use JSON;

our @ISA     = qw(Exporter);
our @EXPORT  = qw();
our $VERSION = 1.00;

use warnings;
use strict;

# 0: url
# 1: username
# 2: password
# 3: data
sub post($$$$) {
    my ($url, $username, $password, $data) = @_;

    my $ua = LWP::UserAgent->new;

    my $json = encode_json($data);
    
    my $request = POST($url);
    $request->authorization_basic($username, $password);
    $request->content_type("application/json");
    $request->content($json);

    my $response = $ua->request($request);

    if ($response->is_success) {
        my $resp = decode_json $response->content;
        if (!defined $resp) {
            return "Cannot decode response JSON"
        }

        if ($resp->{Status} > 0) {
            return $resp->{Message}
        }
        
        return undef
    }

    return $response->status
}

# 0: transaction
# 1: key
# 2: value
# 3: value type
sub addValue($$$;$) {
    $_[0]->{$_[1]}{Value} = $_[2];
    $_[0]->{$_[1]}{Type} = $_[3] if defined $_[3];
}

sub newTransaction() {
    my %r;

    return \%r;
}
    
1;
