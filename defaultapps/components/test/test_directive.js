/*
* Copyright (c) 2015, 2016 Christian Surlykke
*
* This file is part of the refude project. 
* It is distributed under the GPL v2 license.
* Please refer to the LICENSE file for a copy of the license.
*/

appConfigModule.directive('test', function() {
  return {
    scope: true,
    templateUrl: 'components/test/test.html'
  };
});